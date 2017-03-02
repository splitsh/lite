package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

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

type refsFlag []*splitter.Ref

func (r *refsFlag) String() string {
	return fmt.Sprint(*r)
}

func (r *refsFlag) Set(value string) error {
	parts := strings.Split(value, ":")
	from := parts[0]
	to := ""
	if len(parts) > 1 {
		to = parts[1]
	}

	parts = strings.Split(from, "@")
	from = parts[0]
	commit := ""
	if len(parts) > 1 {
		commit = parts[1]
	}

	*r = append(*r, &splitter.Ref{From: from, To: to, Commit: commit})
	return nil
}

var version = "dev"
var prefixes prefixesFlag
var refs refsFlag
var progress, v bool

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Subcommand is required (publish or split)")
		os.Exit(1)
	}

	if os.Args[1] == "publish" {
		run := &splitter.Run{}
		publishCmdFlagSet(run).Parse(os.Args[2:])

		printVersion(v)

		if len(prefixes) == 0 {
			fmt.Fprintln(os.Stderr, "You must provide the directory to split via the --prefix flag")
			os.Exit(1)
		}
		run.Prefixes = []*splitter.Prefix(prefixes)

		run.Refs = []*splitter.Ref(refs)
		if run.Heads {
			if len(run.Refs) > 0 {
				fmt.Fprintln(os.Stderr, "You cannot use the --heads flag with the --ref one")
				os.Exit(1)
			}
		}
		if run.Tags {
			if len(run.Refs) > 0 {
				fmt.Fprintln(os.Stderr, "You cannot use the --tags flag with the --ref one")
				os.Exit(1)
			}
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
	publishCmd.BoolVar(&run.Heads, "heads", false, "Split all heads")
	publishCmd.BoolVar(&run.Tags, "tags", false, "Split all tags")
	publishCmd.Var(&refs, "ref", "Split this reference only (can be used multiple times)")
	publishCmd.StringVar(&run.RemoteURL, "push", "", "Git URL to push splits")
	publishCmd.Var(&prefixes, "prefix", "The directory(ies) to split")
	//publishCmd.BoolVar(&run.Scratch, "scratch", false, "Flush the cache (optional)")
	publishCmd.StringVar(&run.GitVersion, "git", "latest", "Simulate a given version of Git (optional)")
	publishCmd.BoolVar(&run.Progress, "progress", false, "Display splitting progress information")
	publishCmd.BoolVar(&run.Debug, "debug", false, "Display debug information")
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
	splitCmd.StringVar(&config.GitVersion, "git", "latest", "Simulate a given version of Git (optional)")
	splitCmd.BoolVar(&progress, "progress", false, "Show progress bar (optional, cannot be enabled when debug is enabled)")
	splitCmd.BoolVar(&v, "version", false, "Show version")
	splitCmd.StringVar(&config.Path, "path", ".", "The repository path (optional, current directory by default)")
	return splitCmd
}
