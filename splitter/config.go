package splitter

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/libgit2/git2go"
)

// Prefix represents which paths to split
type Prefix struct {
	From string
	To   string
}

// Config represents a split configuration
type Config struct {
	Prefixes   []*Prefix
	Path       string
	Origin     string
	Commit     string
	Target     string
	GitVersion string
	Debug      bool
	Scratch    bool

	// for advanced usage only
	// naming and types subject to change anytime!
	Logger *log.Logger
	DB     *bolt.DB
	RepoMu *sync.Mutex
	Repo   *git.Repository
	Git    int
}

var supportedGitVersions = map[string]int{
	"<1.8.2": 1,
	"<2.8.0": 2,
	"latest": 3,
}

// Split splits a configuration
func (config *Config) Split(result *Result) error {
	state, err := newState(config, result)
	if err != nil {
		return err
	}
	defer state.close()
	return state.split()
}

// SplitWithFeedback splits a configuration with feedback on the CLI
func (config *Config) SplitWithFeedback(progress bool) string {
	result := &Result{}

	var ticker *time.Ticker
	if progress {
		ticker = time.NewTicker(time.Millisecond * 50)
		go func() {
			for range ticker.C {
				msg := fmt.Sprintf("splitting %d commits, %d created", result.Traversed(), result.Created())
				fmt.Fprintf(os.Stderr, "%s\033[K\033[%dD", msg, len(msg))
			}
		}()
	} else {
		fmt.Fprintf(os.Stderr, "splitting\033[%dD", len("splitting"))
	}

	if err := config.Split(result); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if ticker != nil {
		ticker.Stop()
		fmt.Fprint(os.Stderr, "\033[K")
	}

	fmt.Fprintf(os.Stderr, "split %d commits, %d created, in %s", result.Traversed(), result.Created(), result.Duration(time.Millisecond))

	if result.Head() == nil {
		return ""
	}
	return result.Head().String()
}

// Validate validates the configuration
func (config *Config) Validate() error {
	if !git.ReferenceIsValidName(config.Origin) {
		return fmt.Errorf("The origin is not a valid Git reference")
	}

	if config.Target != "" && !git.ReferenceIsValidName(config.Target) {
		return fmt.Errorf("The target is not a valid Git reference")
	}

	git, ok := supportedGitVersions[config.GitVersion]
	if !ok {
		return fmt.Errorf(`The git version can only be one of "<1.8.2", "<2.8.0", or "latest"`)
	}
	config.Git = git

	return nil
}
