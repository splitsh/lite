package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/splitsh/lite/git"
	"github.com/splitsh/lite/splitter"
)

var (
	version = "dev"
)

type prefixesFlag []*splitter.Prefix

func (p *prefixesFlag) String() string {
	return fmt.Sprint(*p)
}

func (p *prefixesFlag) Set(value string) error {
	parts := strings.Split(value, ":")
	from := parts[0]
	to := ""
	if len(parts) > 1 {
		to = parts[1]
	}

	// value must be unique
	for _, prefix := range []*splitter.Prefix(*p) {
		// FIXME: to should be normalized (xxx vs xxx/ for instance)
		if prefix.To == to {
			return fmt.Errorf("Cannot have two prefix splits under the same directory: %s -> %s vs %s -> %s", prefix.From, prefix.To, from, to)
		}
	}

	*p = append(*p, &splitter.Prefix{From: from, To: to})
	return nil
}

var prefixes prefixesFlag
var origin, target, commit, path, gitVersion string
var scratch, debug, quiet, legacy, progress, v, update bool

type publishFlags struct {
	path    string
	update  bool
	noHeads bool
	heads   string
	noTags  bool
	tags    string
	config  string
	debug   bool
	dry     bool

	project *splitter.Project
	repo    *git.Repo
}

func main() {
	splitCmd := flag.NewFlagSet("split", flag.ExitOnError)
	splitCmd.Var(&prefixes, "prefix", "The directory(ies) to split")
	splitCmd.StringVar(&origin, "origin", "HEAD", "The branch to split (optional, defaults to the current one)")
	splitCmd.StringVar(&target, "target", "", "The branch to create when split is finished (optional)")
	splitCmd.StringVar(&commit, "commit", "", "The commit at which to start the split (optional)")
	splitCmd.BoolVar(&scratch, "scratch", false, "Flush the cache (optional)")
	splitCmd.BoolVar(&debug, "debug", false, "Enable the debug mode (optional)")
	splitCmd.BoolVar(&quiet, "quiet", false, "Suppress the output (optional)")
	splitCmd.BoolVar(&legacy, "legacy", false, "[DEPRECATED] Enable the legacy mode for projects migrating from an old version of git subtree split (optional)")
	splitCmd.StringVar(&gitVersion, "git", "latest", "Simulate a given version of Git (optional)")
	splitCmd.BoolVar(&progress, "progress", false, "Show progress bar (optional, cannot be enabled when debug is enabled)")
	splitCmd.BoolVar(&v, "version", false, "Show version")
	splitCmd.StringVar(&path, "path", ".", "The repository path (optional, current directory by default)")

	initCmd := flag.NewFlagSet("init", flag.ExitOnError)
	initCmd.BoolVar(&v, "version", false, "Show version")
	initCmd.StringVar(&path, "path", ".", "The repository path (optional, current directory by default)")

	updateCmd := flag.NewFlagSet("update", flag.ExitOnError)
	updateCmd.BoolVar(&v, "version", false, "Show version")
	updateCmd.StringVar(&path, "path", ".", "The repository path (optional, current directory by default)")

	pf := &publishFlags{}
	publishCmd := flag.NewFlagSet("publish", flag.ExitOnError)
	publishCmd.BoolVar(&pf.update, "update", false, "")
	publishCmd.BoolVar(&pf.noHeads, "no-heads", false, "")
	publishCmd.StringVar(&pf.heads, "heads", "", "")
	publishCmd.StringVar(&pf.config, "config", "", "")
	publishCmd.BoolVar(&pf.noTags, "no-tags", false, "")
	publishCmd.StringVar(&pf.tags, "tags", "", "")
	publishCmd.BoolVar(&pf.debug, "debug", false, "")
	publishCmd.BoolVar(&pf.dry, "dry-run", false, "")
	publishCmd.BoolVar(&v, "version", false, "Show version")
	publishCmd.StringVar(&pf.path, "path", ".", "The repository path (optional, current directory by default)")

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Subcommand is required (init, publish, update, or split)")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "split":
		splitCmd.Parse(os.Args[2:])
	case "init":
		initCmd.Parse(os.Args[2:])
		if initCmd.NArg() != 1 {
			fmt.Fprintln(os.Stderr, "init requires the Git URL to be passed")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Initializing splitsh from \"%s\" in \"%s\"\n", initCmd.Arg(0), path)
		r := &git.Repo{Path: path}
		if err := r.Clone(initCmd.Arg(0)); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	case "update":
		updateCmd.Parse(os.Args[2:])
		fmt.Fprintf(os.Stderr, "Updating repository in \"%s\"\n", path)
		r := &git.Repo{Path: path}
		if err := r.Update(); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	case "publish":
		publishCmd.Parse(os.Args[2:])
		if err := runPublishCmd(pf); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	default:
		// FIXME: deprecated
		splitCmd.Parse(os.Args[1:])
	}

	if v {
		fmt.Fprintf(os.Stderr, "splitsh-lite version %s\n", version)
		os.Exit(0)
	}

	if len(prefixes) == 0 {
		fmt.Fprintln(os.Stderr, "You must provide the directory to split via the --prefix flag")
		os.Exit(1)
	}

	if legacy {
		fmt.Fprintln(os.Stderr, `The --legacy option is deprecated (use --git="<1.8.2" instead)`)
		gitVersion = "<1.8.2"
	}

	config := &splitter.Config{
		Path:       path,
		Origin:     origin,
		Prefixes:   []*splitter.Prefix(prefixes),
		Target:     target,
		Commit:     commit,
		Debug:      debug && !quiet,
		Scratch:    scratch,
		GitVersion: gitVersion,
	}

	sha1 := runSplitCmd(config, progress && !debug && !quiet, quiet)
	if sha1 != "" {
		fmt.Println(sha1)
	}
}

func runSplitCmd(config *splitter.Config, progress, quiet bool) string {
	result := &splitter.Result{}

	var ticker *time.Ticker
	if progress {
		ticker = time.NewTicker(time.Millisecond * 50)
		go func() {
			for range ticker.C {
				fmt.Fprintf(os.Stderr, "%d commits created, %d commits traversed\r", result.Created(), result.Traversed())
			}
		}()
	}

	if err := splitter.Split(config, result); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if ticker != nil {
		ticker.Stop()
	}

	if !quiet {
		fmt.Fprintf(os.Stderr, "%d commits created, %d commits traversed, in %s\n", result.Created(), result.Traversed(), result.Duration(time.Millisecond))
	}

	if result.Head() == nil {
		return ""
	}
	return result.Head().String()
}

func runPublishCmd(pf *publishFlags) error {
	var err error

	config, err := ioutil.ReadFile(pf.config)
	if err != nil {
		return fmt.Errorf("Could not read config file: %s\n", err)
	}

	pf.project, err = splitter.NewProject(config)
	if err != nil {
		return fmt.Errorf("Could read project: %s\n", err)
	}
	pf.repo = &git.Repo{Path: pf.path}

	if pf.update {
		if err := pf.repo.Update(); err != nil {
			return err
		}
	}

	for name, subtree := range pf.project.Subtrees {
		if err := pf.repo.CreateRemote(name, subtree.Target); err != nil {
			return fmt.Errorf("Could create remote: %s\n", err)
		}

		for _, prefix := range subtree.Prefixes {
			fmt.Fprintf(os.Stderr, "Syncing %s -> %s\n", prefix, subtree.Target)
		}

		pf.syncHeads(subtree)
		pf.syncTags(subtree)
	}

	return nil
}

func (pf *publishFlags) syncHeads(subtree *splitter.Subtree) {
	for _, head := range pf.getHeads() {
		fmt.Fprintf(os.Stderr, " syncing head %s", head)
		if !pf.repo.CheckRef("refs/heads/" + head) {
			fmt.Fprintln(os.Stderr, " - skipping, does not exist")
			continue
		}

		config := pf.createConfig(subtree, "refs/heads/"+head)
		sha1 := runSplitCmd(config, progress && !debug && !quiet, !debug)
		if sha1 != "" {
			pf.repo.Push(subtree.Target, sha1, "refs/heads/"+head, pf.dry)
			fmt.Fprintln(os.Stderr, " - pushed")
		} else {
			fmt.Fprintln(os.Stderr, " - empty, not pushed")
		}
	}
}

func (pf *publishFlags) syncTags(subtree *splitter.Subtree) {
	targetTags := pf.repo.RemoteTags(subtree.Target)
NextTag:
	for _, tag := range pf.getTags() {
		fmt.Fprintf(os.Stderr, " syncing tag %s", tag)
		if !pf.repo.CheckRef("refs/tags/" + tag) {
			fmt.Fprintln(os.Stderr, " - skipping, does not exist")
			continue
		}

		for _, t := range targetTags {
			if t == tag {
				fmt.Fprintln(os.Stderr, " - skipping, already synced")
				continue NextTag
			}
		}

		config := pf.createConfig(subtree, "refs/tags/"+tag)
		sha1 := runSplitCmd(config, progress && !debug && !quiet, !debug)
		if sha1 != "" {
			pf.repo.Push(subtree.Target, sha1, "refs/tags/"+tag, pf.dry)
			fmt.Fprintln(os.Stderr, " - pushed")
		} else {
			fmt.Fprintln(os.Stderr, " - empty, not pushed")
		}
	}
}

func (pf *publishFlags) createConfig(subtree *splitter.Subtree, ref string) *splitter.Config {
	prefixes := []*splitter.Prefix{}
	for _, prefix := range subtree.Prefixes {
		parts := strings.Split(prefix, ":")
		from := parts[0]
		to := ""
		if len(parts) > 1 {
			to = parts[1]
		}
		prefixes = append(prefixes, &splitter.Prefix{From: from, To: to})
	}

	return &splitter.Config{
		Path:       pf.path,
		Origin:     ref,
		Prefixes:   prefixes,
		GitVersion: pf.project.GitVersion,
		Debug:      pf.debug,
	}
}

func (pf *publishFlags) getHeads() []string {
	var heads []string

	if pf.noHeads {
		return heads
	}

	if pf.heads != "" {
		return strings.Split(pf.heads, " ")
	}

	return pf.repo.RemoteHeads("origin")
}

func (pf *publishFlags) getTags() []string {
	var tags []string

	if pf.noTags {
		return tags
	}

	if pf.tags != "" {
		return strings.Split(pf.tags, " ")
	}

	return pf.repo.RemoteTags("origin")
}
