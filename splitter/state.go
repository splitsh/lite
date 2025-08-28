package splitter

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	git "github.com/libgit2/git2go/v34"
)

type state struct {
	config       *Config
	origin       string
	repoMu       *sync.Mutex
	repo         *git.Repository
	cache        *cache
	logger       *log.Logger
	simplePrefix string
	result       *Result
}

func newState(config *Config, result *Result) (*state, error) {
	var err error

	// validate config
	if err = config.Validate(); err != nil {
		return nil, err
	}

	state := &state{
		config: config,
		result: result,
		repoMu: config.RepoMu,
		repo:   config.Repo,
		logger: config.Logger,
	}

	if state.repo == nil {
		if state.repo, err = git.OpenRepository(config.Path); err != nil {
			return nil, err
		}
	}

	if state.repoMu == nil {
		state.repoMu = &sync.Mutex{}
	}

	if state.logger == nil {
		state.logger = log.New(os.Stderr, "", log.LstdFlags)
	}

	if state.origin, err = normalizeOrigin(state.repo, config.Origin); err != nil {
		return nil, err
	}

	if state.cache, err = newCache(state.origin, config); err != nil {
		return nil, err
	}

	if config.Debug {
		state.logger.Printf("Splitting %s", state.origin)
		for _, v := range config.Prefixes {
			to := v.To
			if to == "" {
				to = "ROOT"
			}
			state.logger.Printf(`  From "%s" to "%s"`, v.From, to)
			if (len(v.Excludes)) == 0 {
			} else {
				state.logger.Printf(`  Excluding "%s"`, strings.Join(v.Excludes, `", "`))
			}
		}
	}

	if config.Scratch {
		if err := state.flush(); err != nil {
			return nil, err
		}
	}

	// simplePrefix contains the prefix when there is only one
	// with an empty value (target)
	if len(config.Prefixes) == 1 && config.Prefixes[0].To == "" && len(config.Prefixes[0].Excludes) == 0 {
		state.simplePrefix = config.Prefixes[0].From
	}

	return state, nil
}

func (s *state) close() error {
	err := s.cache.close()
	if err != nil {
		return err
	}
	s.repo.Free()
	return nil
}

func (s *state) flush() error {
	if err := s.cache.flush(); err != nil {
		return err
	}

	if s.config.Target != "" {
		branch, err := s.repo.LookupBranch(s.config.Target, git.BranchLocal)
		if err == nil {
			branch.Delete()
			branch.Free()
		}
	}

	return nil
}

func (s *state) split() error {
	startTime := time.Now()
	defer func() {
		s.result.end(startTime)
	}()

	revWalk, err := s.walker()
	if err != nil {
		return fmt.Errorf("impossible to walk the repository: %s", err)
	}
	defer revWalk.Free()

	var iterationErr error
	var lastRev *git.Oid
	err = revWalk.Iterate(func(rev *git.Commit) bool {
		defer rev.Free()
		lastRev = rev.Id()

		if s.config.Debug {
			s.logger.Printf("Processing commit: %s\n", rev.Id().String())
		}

		var newrev *git.Oid
		newrev, err = s.splitRev(rev)
		if err != nil {
			iterationErr = err
			return false
		}

		if newrev != nil {
			s.result.moveHead(newrev)
		}

		return true
	})
	if err != nil {
		return err
	}
	if iterationErr != nil {
		return iterationErr
	}

	if lastRev != nil {
		s.cache.setHead(lastRev)
	}

	return s.updateTarget()
}

func (s *state) walker() (*git.RevWalk, error) {
	revWalk, err := s.repo.Walk()
	if err != nil {
		return nil, fmt.Errorf("impossible to walk the repository: %s", err)
	}

	err = s.pushRevs(revWalk)
	if err != nil {
		return nil, fmt.Errorf("impossible to determine split range: %s", err)
	}

	revWalk.Sorting(git.SortTopological | git.SortReverse)

	return revWalk, nil
}

func (s *state) splitRev(rev *git.Commit) (*git.Oid, error) {
	s.result.incTraversed()

	v := s.cache.get(rev.Id())
	if v != nil {
		if s.config.Debug {
			s.logger.Printf("  prior: %s\n", v.String())
		}
		return v, nil
	}

	var parents []*git.Oid
	var n uint
	for n = 0; n < rev.ParentCount(); n++ {
		parents = append(parents, rev.ParentId(n))
	}

	if s.config.Debug {
		debugMsg := "  parents:"
		for _, parent := range parents {
			debugMsg += fmt.Sprintf(" %s", parent.String())
		}
		s.logger.Print(debugMsg)
	}

	newParents := s.cache.gets(parents)

	if s.config.Debug {
		debugMsg := "  newparents:"
		for _, parent := range newParents {
			debugMsg += fmt.Sprintf(" %s", parent)
		}
		s.logger.Print(debugMsg)
	}

	tree, err := s.subtreeForCommit(rev)
	if err != nil {
		return nil, err
	}

	if nil == tree {
		// should never happen
		return nil, nil
	}
	defer tree.Free()

	if s.config.Debug {
		s.logger.Printf("  tree is: %s\n", tree.Id().String())
	}

	newrev, created, err := s.copyOrSkip(rev, tree, newParents)
	if err != nil {
		return nil, err
	}

	if s.config.Debug {
		s.logger.Printf("  newrev is: %s\n", newrev)
	}

	if created {
		s.result.incCreated()
	}

	s.cache.set(rev.Id(), newrev)

	return newrev, nil
}

func (s *state) subtreeForCommit(commit *git.Commit) (*git.Tree, error) {
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}
	defer tree.Free()

	if s.simplePrefix != "" {
		return s.treeByPath(tree, s.simplePrefix)
	}

	return s.treeByPaths(tree)
}

func (s *state) treeByPath(tree *git.Tree, prefix string) (*git.Tree, error) {
	treeEntry, err := tree.EntryByPath(prefix)
	if err != nil {
		return nil, nil
	}

	if treeEntry.Type != git.ObjectTree {
		// tree is not a tree (a directory for a gitmodule for instance), skip
		return nil, nil
	}

	return s.repo.LookupTree(treeEntry.Id)
}

func (s *state) treeByPaths(tree *git.Tree) (*git.Tree, error) {
	var currentTree, prefixedTree, mergedTree *git.Tree
	for _, prefix := range s.config.Prefixes {
		// splitting
		splitTree, err := s.treeByPath(tree, prefix.From)
		if err != nil {
			return nil, err
		}
		if splitTree == nil {
			continue
		}

		if len(prefix.Excludes) > 0 {
			prunedTree, err := s.pruneTree(splitTree, prefix.Excludes)
			if err != nil {
				return nil, err
			}
			splitTree = prunedTree
		}

		// adding the prefix
		if prefix.To != "" {
			prefixedTree, err = s.addPrefixToTree(splitTree, prefix.To)
			if err != nil {
				return nil, err
			}
		} else {
			prefixedTree = splitTree
		}

		// merging with the current tree
		if currentTree != nil {
			mergedTree, err = s.mergeTrees(currentTree, prefixedTree)
			currentTree.Free()
			prefixedTree.Free()
			if err != nil {
				return nil, err
			}
		} else {
			mergedTree = prefixedTree
		}

		currentTree = mergedTree
	}

	return currentTree, nil
}

func (s *state) mergeTrees(t1, t2 *git.Tree) (*git.Tree, error) {
	index, err := s.repo.MergeTrees(nil, t1, t2, nil)
	if err != nil {
		return nil, err
	}
	defer index.Free()

	if index.HasConflicts() {
		return nil, fmt.Errorf("cannot split as there is a merge conflict between two paths")
	}

	oid, err := index.WriteTreeTo(s.repo)
	if err != nil {
		return nil, err
	}

	return s.repo.LookupTree(oid)
}

func (s *state) addPrefixToTree(tree *git.Tree, prefix string) (*git.Tree, error) {
	treeOid := tree.Id()
	parts := strings.Split(prefix, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		treeBuilder, err := s.repo.TreeBuilder()
		if err != nil {
			return nil, err
		}
		defer treeBuilder.Free()

		err = treeBuilder.Insert(parts[i], treeOid, git.FilemodeTree)
		if err != nil {
			return nil, err
		}

		treeOid, err = treeBuilder.Write()
		if err != nil {
			return nil, err
		}
	}

	prefixedTree, err := s.repo.LookupTree(treeOid)
	if err != nil {
		return nil, err
	}

	return prefixedTree, nil
}

func (s *state) pruneTree(tree *git.Tree, excludes []string) (*git.Tree, error) {
	var err error
	treeBuilder, err := s.repo.TreeBuilder()
	if err != nil {
		return nil, err
	}
	defer treeBuilder.Free()

	err = tree.Walk(func(path string, entry *git.TreeEntry) error {
		// always add files at the root directory
		if entry.Type == git.ObjectBlob {
			if err := treeBuilder.Insert(entry.Name, entry.Id, entry.Filemode); err != nil {
				return err
			}
			return nil
		}

		if entry.Type != git.ObjectTree {
			// should never happen
			return fmt.Errorf("Unexpected entry %s/%s (type %s)", path, entry.Name, entry.Type)
		}

		// exclude directory in excludes
		for _, exclude := range excludes {
			if entry.Name == exclude {
				return git.TreeWalkSkip
			}
		}

		if err := treeBuilder.Insert(entry.Name, entry.Id, git.FilemodeTree); err != nil {
			return err
		}
		return git.TreeWalkSkip
	})

	if err != nil {
		return nil, err
	}

	treeOid, err := treeBuilder.Write()
	if err != nil {
		return nil, err
	}

	return s.repo.LookupTree(treeOid)
}

func (s *state) copyOrSkip(rev *git.Commit, tree *git.Tree, newParents []*git.Oid) (*git.Oid, bool, error) {
	var identical, nonIdentical *git.Oid
	var gotParents []*git.Oid
	var p []*git.Commit
	for _, parent := range newParents {
		ptree, err := s.topTreeForCommit(parent)
		if err != nil {
			return nil, false, err
		}
		if nil == ptree {
			continue
		}

		if ptree.Cmp(tree.Id()) == 0 {
			// an identical parent could be used in place of this rev.
			identical = parent
		} else {
			nonIdentical = parent
		}

		// sometimes both old parents map to the same newparent
		// eliminate duplicates
		isNew := true
		for _, gp := range gotParents {
			if gp.Cmp(parent) == 0 {
				isNew = false
				break
			}
		}

		if isNew {
			gotParents = append(gotParents, parent)
			commit, err := s.repo.LookupCommit(parent)
			if err != nil {
				return nil, false, err
			}
			defer commit.Free()
			p = append(p, commit)
		}
	}

	copyCommit := false
	if s.config.Git > 2 && nil != identical && nil != nonIdentical {
		revWalk, err := s.repo.Walk()
		if err != nil {
			return nil, false, fmt.Errorf("impossible to walk the repository: %s", err)
		}

		s.repoMu.Lock()
		defer s.repoMu.Unlock()

		err = revWalk.PushRange(fmt.Sprintf("%s..%s", identical, nonIdentical))
		if err != nil {
			return nil, false, fmt.Errorf("impossible to determine split range: %s", err)
		}

		err = revWalk.Iterate(func(rev *git.Commit) bool {
			// we need to preserve history along the other branch
			copyCommit = true
			return false
		})
		if err != nil {
			return nil, false, err
		}

		revWalk.Free()
	}

	if nil != identical && !copyCommit {
		return identical, false, nil
	}

	commit, err := s.copyCommit(rev, tree, p)
	if err != nil {
		return nil, false, err
	}

	return commit, true, nil
}

func (s *state) topTreeForCommit(sha *git.Oid) (*git.Oid, error) {
	commit, err := s.repo.LookupCommit(sha)
	if err != nil {
		return nil, err
	}
	defer commit.Free()

	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}
	defer tree.Free()

	return tree.Id(), nil
}

func (s *state) copyCommit(rev *git.Commit, tree *git.Tree, parents []*git.Commit) (*git.Oid, error) {
	if s.config.Debug {
		parentStrs := make([]string, len(parents))
		for i, parent := range parents {
			parentStrs[i] = parent.Id().String()
		}
		s.logger.Printf("  copy commit \"%s\" \"%s\" \"%s\"\n", rev.Id().String(), tree.Id().String(), strings.Join(parentStrs, " "))
	}

	message := rev.RawMessage()
	if s.config.Git == 1 {
		message = s.legacyMessage(rev)
	}

	author := rev.Author()
	if author.Email == "" {
		author.Email = "nobody@example.com"
	}

	committer := rev.Committer()
	if committer.Email == "" {
		committer.Email = "nobody@example.com"
	}

	oid, err := s.repo.CreateCommit("", author, committer, message, tree, parents...)
	if err != nil {
		return nil, err
	}

	return oid, nil
}

func (s *state) updateTarget() error {
	if s.config.Target == "" {
		return nil
	}

	if nil == s.result.Head() {
		return fmt.Errorf("unable to create branch %s as it is empty (no commits were split)", s.config.Target)
	}

	obj, ref, err := s.repo.RevparseExt(s.config.Target)
	if obj != nil {
		obj.Free()
	}
	if err != nil {
		ref, err = s.repo.References.Create(s.config.Target, s.result.Head(), false, "subtree split")
		if err != nil {
			return err
		}
		ref.Free()
	} else {
		defer ref.Free()
		ref.SetTarget(s.result.Head(), "subtree split")
	}

	return nil
}

func (s *state) legacyMessage(rev *git.Commit) string {
	subject, body := SplitMessage(rev.Message())
	return subject + "\n\n" + body
}

// pushRevs sets the range to split
func (s *state) pushRevs(revWalk *git.RevWalk) error {
	s.repoMu.Lock()
	defer s.repoMu.Unlock()

	var start *git.Oid
	start = s.cache.getHead()
	if start != nil {
		s.result.moveHead(s.cache.get(start))
		// FIXME: CHECK that this is an ancestor of the branch?
		return revWalk.PushRange(fmt.Sprintf("%s..%s", start, s.origin))
	}

	// find the latest split sha1 if any on origin
	var err error
	if s.config.Commit != "" {
		start, err = git.NewOid(s.config.Commit)
		if err != nil {
			return err
		}
		s.result.moveHead(s.cache.get(start))
		return revWalk.PushRange(fmt.Sprintf("%s^..%s", start, s.origin))
	}

	branch, err := s.repo.RevparseSingle(s.origin)
	if err != nil {
		return err
	}

	return revWalk.Push(branch.Id())
}
