package splitter

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/splitsh/lite/git"
)

// Run represents a run from the CLI
type Run struct {
	Path     string
	NoUpdate bool
	NoHeads  bool
	Heads    string
	NoTags   bool
	Tags     string
	Config   string
	Debug    bool
	Progress bool
	DryRun   bool

	repo *git.Repo
}

// Sync synchronizes branches and tags
func (r *Run) Sync() error {
	project, err := r.createProject()
	if err != nil {
		return err
	}
	r.repo = &git.Repo{Path: r.Path}

	if !r.NoUpdate {
		fmt.Fprintln(os.Stderr, "Fetching changes from origin")
		if err := r.repo.Update(); err != nil {
			return err
		}
	}

	for name, subtree := range project.Subtrees {
		if err := r.repo.CreateRemote(name, subtree.Target); err != nil {
			return fmt.Errorf("Could create remote: %s\n", err)
		}

		for _, prefix := range subtree.Prefixes {
			fmt.Fprintf(os.Stderr, "Syncing %s -> %s\n", prefix, subtree.Target)
		}

		r.syncHeads(project, subtree)
		r.syncTags(project, subtree)
	}
	return nil
}

func (r *Run) syncHeads(project *Project, subtree *Subtree) {
	for _, head := range r.getHeads() {
		fmt.Fprintf(os.Stderr, "  Head %s", head)
		if !r.repo.CheckRef("refs/heads/" + head) {
			fmt.Fprintln(os.Stderr, " > skipping, does not exist")
			continue
		}

		fmt.Fprint(os.Stderr, " > ")
		config := r.createConfig(project, subtree, "refs/heads/"+head)
		if sha1 := config.SplitWithFeedback(r.Progress); sha1 != "" {
			fmt.Fprint(os.Stderr, " > pushing")
			r.repo.Push(subtree.Target, sha1, "refs/heads/"+head, r.DryRun)
			fmt.Fprintln(os.Stderr, " > pushed")
		} else {
			fmt.Fprintln(os.Stderr, " > empty, not pushed")
		}
	}
}

func (r *Run) syncTags(project *Project, subtree *Subtree) {
	targetTags := r.repo.RemoteTags(subtree.Target)
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
		config := r.createConfig(project, subtree, "refs/tags/"+tag)
		if sha1 := config.SplitWithFeedback(r.Progress); sha1 != "" {
			fmt.Fprint(os.Stderr, " > pushing")
			r.repo.Push(subtree.Target, sha1, "refs/tags/"+tag, r.DryRun)
			fmt.Fprintln(os.Stderr, " > pushed")
		} else {
			fmt.Fprintln(os.Stderr, " > empty, not pushed")
		}
	}
}

func (r *Run) createConfig(project *Project, subtree *Subtree, ref string) *Config {
	prefixes := []*Prefix{}
	for _, prefix := range subtree.Prefixes {
		parts := strings.Split(prefix, ":")
		from := parts[0]
		to := ""
		if len(parts) > 1 {
			to = parts[1]
		}
		prefixes = append(prefixes, &Prefix{From: from, To: to})
	}

	return &Config{
		Path:       r.Path,
		Origin:     ref,
		Prefixes:   prefixes,
		GitVersion: project.GitVersion,
		Debug:      r.Debug,
	}
}

func (r *Run) getHeads() []string {
	var heads []string

	if r.NoHeads {
		return heads
	}

	if r.Heads != "" {
		return strings.Split(r.Heads, " ")
	}

	return r.repo.RemoteHeads("origin")
}

func (r *Run) getTags() []string {
	var tags []string

	if r.NoTags {
		return tags
	}

	if r.Tags != "" {
		return strings.Split(r.Tags, " ")
	}

	return r.repo.RemoteTags("origin")
}

func (r *Run) createProject() (*Project, error) {
	config, err := ioutil.ReadFile(r.Config)
	if err != nil {
		return nil, fmt.Errorf("Could not read config file: %s\n", err)
	}

	project, err := NewProject(config)
	if err != nil {
		return nil, fmt.Errorf("Could read project: %s\n", err)
	}

	return project, nil
}
