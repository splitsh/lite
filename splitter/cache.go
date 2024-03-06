package splitter

import (
	"crypto/sha1"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"time"

	git "github.com/libgit2/git2go/v34"
	bolt "go.etcd.io/bbolt"
)

type cache struct {
	key    []byte
	branch string
	db     *bolt.DB
	data   map[string][]byte
}

func newCache(branch string, config *Config) (*cache, error) {
	var err error
	db := config.DB
	if db == nil {
		db, err = bolt.Open(filepath.Join(GitDirectory(config.Path), "splitsh.db"), 0644, &bolt.Options{Timeout: 5 * time.Second})
		if err != nil {
			return nil, err
		}
	}

	c := &cache{
		db:     db,
		branch: branch,
		key:    key(config),
		data:   make(map[string][]byte),
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err1 := tx.CreateBucketIfNotExists(c.key)
		return err1
	})
	if err != nil {
		return nil, fmt.Errorf("Impossible to create bucket: %s", err)
	}

	return c, nil
}

func (c *cache) close() error {
	err := c.db.Update(func(tx *bolt.Tx) error {
		for k, v := range c.data {
			if err := tx.Bucket(c.key).Put([]byte(k), v); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return c.db.Close()
}

func key(config *Config) []byte {
	h := sha1.New()
	if config.Commit != "" {
		io.WriteString(h, config.Commit)
	} else {
		// value does not matter, should just be always the same
		io.WriteString(h, "oldest")
	}

	io.WriteString(h, strconv.Itoa(config.Git))

	for _, prefix := range config.Prefixes {
		io.WriteString(h, prefix.From)
		io.WriteString(h, prefix.To)
	}

	return h.Sum(nil)
}

func (c *cache) setHead(head *git.Oid) {
	c.data["head/"+c.branch] = head[0:20]
}

func (c *cache) getHead() *git.Oid {
	if head, ok := c.data["head"+c.branch]; ok {
		return git.NewOidFromBytes(head)
	}

	var oid *git.Oid
	c.db.View(func(tx *bolt.Tx) error {
		result := tx.Bucket(c.key).Get([]byte("head/" + c.branch))
		if result != nil {
			c.data["head/"+c.branch] = result
			oid = git.NewOidFromBytes(result)
		}
		return nil
	})
	return oid
}

func (c *cache) get(rev *git.Oid) *git.Oid {
	if v, ok := c.data[string(rev[0:20])]; ok {
		return git.NewOidFromBytes(v)
	}

	var oid *git.Oid
	c.db.View(func(tx *bolt.Tx) error {
		result := tx.Bucket(c.key).Get(rev[0:20])
		if result != nil {
			c.data[string(rev[0:20])] = result
			oid = git.NewOidFromBytes(result)
		}
		return nil
	})
	return oid
}

func (c *cache) set(rev, newrev *git.Oid, created bool) {
	c.data[string(rev[0:20])] = newrev[0:20]
	postfix := "/newest"
	if created {
		postfix = "/oldest"
	}
	c.data[string(append(newrev[0:20], []byte(postfix)...))] = rev[0:20]
}

func (c *cache) gets(commits []*git.Oid) []*git.Oid {
	var oids []*git.Oid
	c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(c.key)
		for _, commit := range commits {
			result := c.data[string(commit[0:20])]
			if result != nil {
				oids = append(oids, git.NewOidFromBytes(result))
			} else {
				result := b.Get(commit[0:20])
				if result != nil {
					oids = append(oids, git.NewOidFromBytes(result))
				}
			}
		}
		return nil
	})
	return oids
}

func (c *cache) flush() error {
	return c.db.Update(func(tx *bolt.Tx) error {
		if tx.Bucket(c.key) != nil {
			err := tx.DeleteBucket(c.key)
			if err != nil {
				return err
			}

			_, err = tx.CreateBucketIfNotExists(c.key)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
