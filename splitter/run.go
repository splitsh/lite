package splitter

import (
	"fmt"
	"os"

	"github.com/splitsh/lite/git"
)

// Ref represents a refence to split
type Ref struct {
	From   string
	To     string
	Commit string
}

func (r *Ref) String() string {
	s := r.From
	if r.Commit != "" {
		s = fmt.Sprintf("%s@%s", s, r.Commit)
	}
	if r.To != "" {
		s = fmt.Sprintf("%s:%s", s, r.To)
	}
	return s
}

// Run represents a run from the CLI
type Run struct {
	Path       string
	NoUpdate   bool
	Refs       []*Ref
	Prefixes   []*Prefix
	Heads      bool
	Tags       bool
	Debug      bool
	Progress   bool
	DryRun     bool
	RemoteURL  string
	GitVersion string

	repo *git.Repo
}

// Sync synchronizes branches and tags
func (r *Run) Sync() error {
	r.repo = &git.Repo{Path: r.Path}

	if r.Heads {
		for _, ref := range r.repo.RemoteRefs("origin", "refs/heads/") {
			r.Refs = append(r.Refs, &Ref{From: ref})
		}
	}
	if r.Tags {
		for _, ref := range r.repo.RemoteRefs("origin", "refs/tags/") {
			r.Refs = append(r.Refs, &Ref{From: ref})
		}
	}

	for _, ref := range r.Refs {
		fmt.Println(ref)
	}
	os.Exit(0)

	if !r.NoUpdate {
		fmt.Fprintln(os.Stderr, "Fetching changes from origin")
		if err := r.repo.Update(); err != nil {
			return err
		}
	}

	if r.RemoteURL != "" {
		if err := r.repo.CreateRemote(r.RemoteURL); err != nil {
			return fmt.Errorf("Could create remote: %s\n", err)
		}
	}

	for _, prefix := range r.Prefixes {
		if r.RemoteURL != "" {
			fmt.Fprintf(os.Stderr, " %s -> %s\n", prefix, r.RemoteURL)
		} else {
			fmt.Fprintf(os.Stderr, " %s\n", prefix)
		}
	}

	r.syncHeads()
	r.syncTags()
	return nil
}

func (r *Run) syncHeads() {
	for _, head := range r.getHeads() {
		fmt.Fprintf(os.Stderr, "  Head %s", head)
		if !r.repo.CheckRef("refs/heads/" + head) {
			fmt.Fprintln(os.Stderr, " > skipping, does not exist")
			continue
		}

		fmt.Fprint(os.Stderr, " > ")
		config := r.createConfig("refs/heads/" + head)
		if sha1 := config.SplitWithFeedback(r.Progress); sha1 != "" {
			fmt.Fprint(os.Stderr, " > pushing")
			r.repo.Push(r.RemoteURL, sha1, "refs/heads/"+head, r.DryRun)
			fmt.Fprintln(os.Stderr, " > pushed")
		} else {
			fmt.Fprintln(os.Stderr, " > empty, not pushed")
		}
	}
}

func (r *Run) syncTags() {
	targetTags := r.repo.RemoteRefs(r.RemoteURL, "refs/tags/")
NextTag:
	for _, tag := range r.getTags() {
		fmt.Fprintf(os.Stderr, "  Tag %s", tag)
		if !r.repo.CheckRef("refs/tags/" + tag) {
			fmt.Fprintln(os.Stderr, " > skipping, does not exist")
			continue
		}

		for _, t := range targetTags {
			if t == tag {
				fmt.Fprintln(os.Stderr, " > skipping, already synced")
				continue NextTag
			}
		}

		fmt.Fprint(os.Stderr, " > ")
		config := r.createConfig("refs/tags/" + tag)
		if sha1 := config.SplitWithFeedback(r.Progress); sha1 != "" {
			fmt.Fprint(os.Stderr, " > pushing")
			r.repo.Push(r.RemoteURL, sha1, "refs/tags/"+tag, r.DryRun)
			fmt.Fprintln(os.Stderr, " > pushed")
		} else {
			fmt.Fprintln(os.Stderr, " > empty, not pushed")
		}
	}
}

func (r *Run) createConfig(ref string) *Config {
	return &Config{
		Path:       r.Path,
		Origin:     ref,
		Prefixes:   r.Prefixes,
		GitVersion: r.GitVersion,
		Debug:      r.Debug,
	}
}

func (r *Run) getHeads() []string {
	return []string{}
}

func (r *Run) getTags() []string {
	return []string{}
}
