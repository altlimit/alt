package cmd

import (
	"fmt"
	"strings"
)

// ParseRepoArg parses "user/repo[@tag]" into repo ("user/repo") and optional tag.
func ParseRepoArg(arg string) (repo, tag string, err error) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		return "", "", fmt.Errorf("repository argument cannot be empty\n\nExpected format: user/repo or user/repo@tag\nExamples:\n  alt install altlimit/sitegen\n  alt install altlimit/taskr@v0.1.7")
	}

	// Split on @ for tag.
	parts := strings.SplitN(arg, "@", 2)
	repo = parts[0]
	if len(parts) == 2 {
		tag = parts[1]
		if tag == "" {
			return "", "", fmt.Errorf("tag cannot be empty when using @\n\nExpected format: user/repo@tag\nExample: alt install altlimit/taskr@v0.1.7")
		}
	}

	// Validate repo format.
	repoParts := strings.Split(repo, "/")
	if len(repoParts) != 2 || repoParts[0] == "" || repoParts[1] == "" {
		return "", "", fmt.Errorf("invalid repository format: %q\n\nExpected format: user/repo\nExamples:\n  altlimit/sitegen\n  altlimit/taskr\n  altlimit/altclaw", arg)
	}

	return repo, tag, nil
}

// SplitRepo splits "user/repo" into owner and name.
func SplitRepo(repo string) (owner, name string) {
	parts := strings.SplitN(repo, "/", 2)
	return parts[0], parts[1]
}
