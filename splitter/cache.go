package splitter

import (
	"crypto/sha1"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
	git "github.com/libgit2/git2go/v34"
)

type cache struct {
	key    []byte
	branch string
	db     *bolt.DB
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

func (c *cache) setHead(head *git.Oid) error {
	return c.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(c.key).Put([]byte("head/"+c.branch), head[0:20])
	})
}

func (c *cache) getHead() *git.Oid {
	var oid *git.Oid
	c.db.View(func(tx *bolt.Tx) error {
		result := tx.Bucket(c.key).Get([]byte("head/" + c.branch))
		if result != nil {
			oid = git.NewOidFromBytes(result)
		}
		return nil
	})
	return oid
}

func (c *cache) get(rev *git.Oid) *git.Oid {
	var oid *git.Oid
	c.db.View(func(tx *bolt.Tx) error {
		result := tx.Bucket(c.key).Get(rev[0:20])
		if result != nil {
			oid = git.NewOidFromBytes(result)
		}
		return nil
	})
	return oid
}

func (c *cache) set(rev, newrev *git.Oid, created bool) error {
	return c.db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket(c.key).Put(rev[0:20], newrev[0:20])
		if err != nil {
			return err
		}

		postfix := "/newest"
		if created {
			postfix = "/oldest"
		}

		key := append(newrev[0:20], []byte(postfix)...)
		return tx.Bucket(c.key).Put(key, rev[0:20])
	})
}

func (c *cache) gets(commits []*git.Oid) []*git.Oid {
	var oids []*git.Oid

	c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(c.key)
		for _, commit := range commits {
			result := b.Get(commit[0:20])
			if result != nil {
				oids = append(oids, git.NewOidFromBytes(result))
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
