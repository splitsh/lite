package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

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

	run := &splitter.Run{}
	publishCmd := flag.NewFlagSet("publish", flag.ExitOnError)
	publishCmd.BoolVar(&run.Update, "update", false, "")
	publishCmd.BoolVar(&run.NoHeads, "no-heads", false, "")
	publishCmd.StringVar(&run.Heads, "heads", "", "")
	publishCmd.StringVar(&run.Config, "config", "", "")
	publishCmd.BoolVar(&run.NoTags, "no-tags", false, "")
	publishCmd.StringVar(&run.Tags, "tags", "", "")
	publishCmd.BoolVar(&run.Debug, "debug", false, "")
	publishCmd.BoolVar(&run.DryRun, "dry-run", false, "")
	publishCmd.BoolVar(&run.Progress, "progress", false, "")
	publishCmd.BoolVar(&v, "version", false, "Show version")
	publishCmd.StringVar(&run.Path, "path", ".", "The repository path (optional, current directory by default)")

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Subcommand is required (init, publish, update, or split)")
		os.Exit(1)
	}

	switch os.Args[1] {
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
		if err := run.Sync(); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	case "split":
		splitCmd.Parse(os.Args[2:])
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

	/*
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
	*/
}
