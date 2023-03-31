package splitter

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	git "github.com/libgit2/git2go/v34"
)

var messageNormalizer = regexp.MustCompile(`\s*\r?\n`)

// GitDirectory returns the .git directory for a given directory
func GitDirectory(path string) string {
	gitPath := filepath.Join(path, ".git")
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		// this might be a bare repo
		return path
	}

	return gitPath
}

// SplitMessage splits a git message
func SplitMessage(message string) (string, string) {
	// we split the message at \n\n or \r\n\r\n
	var subject, body string
	found := false
	for i := 0; i+4 <= len(message); i++ {
		if message[i] == '\n' && message[i+1] == '\n' {
			subject = message[0:i]
			body = message[i+2:]
			found = true
			break
		} else if message[i] == '\r' && message[i+1] == '\n' && message[i+2] == '\r' && message[i+3] == '\n' {
			subject = message[0:i]
			body = message[i+4:]
			found = true
			break
		}
	}

	if !found {
		subject = message
		body = ""
	}

	// normalize \r\n and whitespaces
	subject = messageNormalizer.ReplaceAllLiteralString(subject, " ")

	// remove spaces at the end of the subject
	subject = strings.TrimRight(subject, " ")
	body = strings.TrimLeft(body, "\r\n")
	return subject, body
}

func normalizeOriginBranch(repo *git.Repository, origin string) (string, error) {
	if origin == "" {
		origin = "HEAD"
	}

	obj, ref, err := repo.RevparseExt(origin)
	if err != nil {
		return "", fmt.Errorf("Bad revision for origin: %s", err)
	}
	if obj != nil {
		obj.Free()
	}
	defer ref.Free()

	return ref.Name(), nil
}
