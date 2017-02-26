package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/splitsh/lite/git"
	"github.com/splitsh/lite/splitter"
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

var version = "dev"
var prefixes prefixesFlag
var origin, target, commit, path, gitVersion string
var scratch, debug, legacy, progress, v, update bool

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
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Subcommand is required (publish or split)")
		os.Exit(1)
	}

	if os.Args[1] == "publish" {
		run := &splitter.Run{}
		publishCmdFlagSet(run).Parse(os.Args[2:])
		printVersion(v)
		if run.Config == "" {
			fmt.Fprintln(os.Stderr, "You must provide the configuration via the --config flag")
			os.Exit(1)
		}
		if err := run.Sync(); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	} else if os.Args[1] == "split" {
		config := &splitter.Config{}
		splitCmdFlagSet(config).Parse(os.Args[2:])
		printVersion(v)
		runSplitCmd(config)
	} else {
		fmt.Fprintln(os.Stderr, "Unknown command, should be publish or split")
		os.Exit(1)
	}
}

func runSplitCmd(config *splitter.Config) {
	if len(prefixes) == 0 {
		fmt.Fprintln(os.Stderr, "You must provide the directory to split via the --prefix flag")
		os.Exit(1)
	}

	if legacy {
		fmt.Fprintln(os.Stderr, `The --legacy option is deprecated (use --git="<1.8.2" instead)`)
		gitVersion = "<1.8.2"
	}

	config.Prefixes = []*splitter.Prefix(prefixes)
	sha1 := config.SplitWithFeedback(progress && !config.Debug)
	fmt.Fprintln(os.Stderr, "")
	if sha1 != "" {
		fmt.Println(sha1)
	}
}

func printVersion(v bool) {
	if v {
		fmt.Fprintf(os.Stderr, "splitsh-lite version %s\n", version)
		os.Exit(0)
	}
}

func publishCmdFlagSet(run *splitter.Run) *flag.FlagSet {
	publishCmd := flag.NewFlagSet("publish", flag.ExitOnError)
	publishCmd.BoolVar(&run.NoUpdate, "no-update", false, "Do not fetch origin changes")
	publishCmd.BoolVar(&run.NoHeads, "no-heads", false, "Do not publish any heads")
	publishCmd.StringVar(&run.Heads, "heads", "", "Only publish for listed heads instead of all heads")
	publishCmd.StringVar(&run.Config, "config", "", "JSON file path for the configuration")
	publishCmd.BoolVar(&run.NoTags, "no-tags", false, "Do not publish any tags")
	publishCmd.StringVar(&run.Tags, "tags", "", "Only publish for listed tags instead of all tags")
	publishCmd.BoolVar(&run.Debug, "debug", false, "Display debug information")
	publishCmd.BoolVar(&run.DryRun, "dry-run", false, "Do everything except actually send the updates")
	publishCmd.BoolVar(&run.Progress, "progress", false, "Display splitting progress information")
	publishCmd.BoolVar(&v, "version", false, "Show version")
	publishCmd.StringVar(&run.Path, "path", ".", "The repository path (optional, current directory by default)")
	return publishCmd
}

func splitCmdFlagSet(config *splitter.Config) *flag.FlagSet {
	splitCmd := flag.NewFlagSet("split", flag.ExitOnError)
	splitCmd.Var(&prefixes, "prefix", "The directory(ies) to split")
	splitCmd.StringVar(&config.Origin, "origin", "HEAD", "The branch to split (optional, defaults to the current one)")
	splitCmd.StringVar(&config.Target, "target", "", "The branch to create when split is finished (optional)")
	splitCmd.StringVar(&config.Commit, "commit", "", "The commit at which to start the split (optional)")
	splitCmd.BoolVar(&config.Scratch, "scratch", false, "Flush the cache (optional)")
	splitCmd.BoolVar(&config.Debug, "debug", false, "Enable the debug mode (optional)")
	splitCmd.BoolVar(&legacy, "legacy", false, "[DEPRECATED] Enable the legacy mode for projects migrating from an old version of git subtree split (optional)")
	splitCmd.StringVar(&config.GitVersion, "git", "latest", "Simulate a given version of Git (optional)")
	splitCmd.BoolVar(&progress, "progress", false, "Show progress bar (optional, cannot be enabled when debug is enabled)")
	splitCmd.BoolVar(&v, "version", false, "Show version")
	splitCmd.StringVar(&config.Path, "path", ".", "The repository path (optional, current directory by default)")
	return splitCmd
}
