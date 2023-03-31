package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/deadworoz/splitsh-lite/splitter"
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
var scratch, debug, quiet, legacy, progress, v bool

func init() {
	flag.Var(&prefixes, "prefix", "The directory(ies) to split")
	flag.StringVar(&origin, "origin", "HEAD", "The branch to split (optional, defaults to the current one)")
	flag.StringVar(&target, "target", "", "The branch to create when split is finished (optional)")
	flag.StringVar(&commit, "commit", "", "The commit at which to start the split (optional)")
	flag.StringVar(&path, "path", ".", "The repository path (optional, current directory by default)")
	flag.BoolVar(&scratch, "scratch", false, "Flush the cache (optional)")
	flag.BoolVar(&debug, "debug", false, "Enable the debug mode (optional)")
	flag.BoolVar(&quiet, "quiet", false, "[DEPRECATED] Suppress the output (optional)")
	flag.BoolVar(&legacy, "legacy", false, "[DEPRECATED] Enable the legacy mode for projects migrating from an old version of git subtree split (optional)")
	flag.StringVar(&gitVersion, "git", "latest", "Simulate a given version of Git (optional)")
	flag.BoolVar(&progress, "progress", false, "Show progress bar (optional, cannot be enabled when debug is enabled)")
	flag.BoolVar(&v, "version", false, "Show version")
}

func main() {
	flag.Parse()

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

	if quiet {
		fmt.Fprintln(os.Stderr, `The --quiet option is deprecated (append 2>/dev/null to the command instead)`)
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

	result := &splitter.Result{}

	var ticker *time.Ticker
	if progress && !debug && !quiet {
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

	if result.Head() != nil {
		fmt.Println(result.Head().String())
	}
}
