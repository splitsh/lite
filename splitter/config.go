package splitter

import (
	"fmt"
	"log"
	"sync"

	"github.com/boltdb/bolt"
	git "github.com/libgit2/git2go/v34"
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
