package splitter

import (
	"sync"
	"time"

	git "github.com/libgit2/git2go/v34"
)

// Result represents the outcome of a split
type Result struct {
	mu        sync.RWMutex
	traversed int
	created   int
	head      *git.Oid
	duration  time.Duration
}

// NewResult returns a pre-populated result
func NewResult(duration time.Duration, traversed, created int) *Result {
	return &Result{
		duration:  duration,
		traversed: traversed,
		created:   created,
	}
}

// Traversed returns the number of commits traversed during the split
func (r *Result) Traversed() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.traversed
}

// Created returns the number of created commits
func (r *Result) Created() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.created
}

// Duration returns the current duration of the split
func (r *Result) Duration(precision time.Duration) time.Duration {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return roundDuration(r.duration, precision)
}

// Head returns the latest split sha1
func (r *Result) Head() *git.Oid {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.head
}

func (r *Result) moveHead(oid *git.Oid) {
	r.mu.Lock()
	r.head = oid
	r.mu.Unlock()
}

func (r *Result) incCreated() {
	r.mu.Lock()
	r.created++
	r.mu.Unlock()
}

func (r *Result) incTraversed() {
	r.mu.Lock()
	r.traversed++
	r.mu.Unlock()
}

func (r *Result) end(start time.Time) {
	r.mu.Lock()
	r.duration = time.Since(start)
	r.mu.Unlock()
}

// roundDuration rounds a duration to a given precision (use roundDuration(d, 10*time.Second) to get a 10s precision fe)
func roundDuration(d, r time.Duration) time.Duration {
	if r <= 0 {
		return d
	}
	neg := d < 0
	if neg {
		d = -d
	}
	if m := d % r; m+m < r {
		d = d - m
	} else {
		d = d + r - m
	}
	if neg {
		return -d
	}
	return d
}
