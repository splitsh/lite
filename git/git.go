package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Repo represents a Git repository
type Repo struct {
	Path string

	remoteRefs map[string][]string
}

// CreateRemote registers a remote if it is not already registered
func (r *Repo) CreateRemote(name, URL string) error {
	cmd := exec.Command("git", "remote")
	cmd.Dir = r.Path
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	for _, n := range strings.Split(string(output), "\n") {
		if n == name {
			return nil
		}
	}

	cmd = exec.Command("git", "remote", "add", name, URL)
	cmd.Dir = r.Path
	return cmd.Run()
}

// RemoteRefs returns the current remotes
func (r *Repo) RemoteRefs(remote string) []string {
	if refs, ok := r.remoteRefs[remote]; ok {
		return refs
	}

	cmd := exec.Command("git", "ls-remote", remote)
	cmd.Dir = r.Path
	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could get remote refs on \"%s\": %s\n", remote, err)
		return []string{}
	}

	if r.remoteRefs == nil {
		r.remoteRefs = make(map[string][]string)
	}
	//	r.remoteRefs[remote] = []string{}
	for _, line := range strings.Split(string(output), "\n") {
		parts := strings.Split(line, "\t")
		if len(parts) > 1 {
			r.remoteRefs[remote] = append(r.remoteRefs[remote], parts[1])
		}
	}

	return r.remoteRefs[remote]
}

// RemoteTags returns tags defined on origin
func (r *Repo) RemoteTags(remote string) []string {
	tags := []string{}
	for _, ref := range r.RemoteRefs(remote) {
		if !strings.Contains(ref, "^{}") && strings.HasPrefix(ref, "refs/tags/") {
			tags = append(tags, strings.TrimPrefix(ref, "refs/tags/"))
		}
	}
	return tags
}

// RemoteHeads returns heads defined on origin
func (r *Repo) RemoteHeads(remote string) []string {
	heads := []string{}
	for _, ref := range r.RemoteRefs(remote) {
		if strings.HasPrefix(ref, "refs/heads/") {
			heads = append(heads, strings.TrimPrefix(ref, "refs/heads/"))
		}
	}
	return heads
}

// CheckRef returns true if the head exists
func (r *Repo) CheckRef(head string) bool {
	cmd := exec.Command("git", "show-ref", "--quiet", "--verify", "--", head)
	cmd.Dir = r.Path
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// Clone clones a given URL
func (r *Repo) Clone(URL string) error {
	cmd := exec.Command("git", "clone", "--bare", "-q", URL, r.Path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Could not clone repository: %s\n", err)
	}
	return nil
}

// Update fetches changes on the origin
func (r *Repo) Update() error {
	cmd := exec.Command("git", "fetch", "-q", "-t", "origin")
	cmd.Dir = r.Path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Could not update repository: %s\n", err)
	}
	return nil
}

// Push pushes a branch to a remote
func (r *Repo) Push(remote, sha1, head string, dry bool) error {
	args := []string{"push"}
	if dry {
		args = append(args, "--dry-run")
	}
	args = append(args, []string{"-q", "--force", remote, sha1 + ":" + head}...)
	cmd := exec.Command("git", args...)
	cmd.Dir = r.Path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Could not push remote \"%s\": %s\n", remote, err)
	}
	return nil
}
