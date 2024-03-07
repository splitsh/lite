package splitter

import (
	"fmt"
	"log"
	"sync"

	git "github.com/libgit2/git2go/v34"
	bolt "go.etcd.io/bbolt"
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
func Split(config *Config, result *Result) error {
	state, err := newState(config, result)
	if err != nil {
		return err
	}
	defer state.close()
	return state.split()
}

// Validate validates the configuration
func (config *Config) Validate() error {
	ok, err := git.ReferenceNameIsValid(config.Origin)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("the origin is not a valid Git reference")
	}

	ok, err = git.ReferenceNameIsValid(config.Target)
	if err != nil {
		return err
	}
	if config.Target != "" && !ok {
		return fmt.Errorf("the target is not a valid Git reference")
	}

	git, ok := supportedGitVersions[config.GitVersion]
	if !ok {
		return fmt.Errorf(`the git version can only be one of "<1.8.2", "<2.8.0", or "latest"`)
	}
	config.Git = git

	return nil
}
