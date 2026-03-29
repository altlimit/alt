package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/altlimit/alt/internal/manifest"
	"github.com/altlimit/alt/internal/platform"
	"github.com/altlimit/alt/internal/shim"
)

// Purge handles `alt purge <repo|user> [...]`.
// Removes binaries, history, shims — the nuclear option.
func Purge(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: alt purge <user/repo|user> [...]\n\nExamples:\n  alt purge altlimit/sitegen                Remove a specific tool\n  alt purge altlimit                        Remove all tools by altlimit\n  alt purge altlimit/sitegen altlimit/taskr  Remove multiple tools")
	}

	m, err := manifest.Load()
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	var entries []manifest.Entry
	for _, arg := range args {
		matched := m.MatchEntries(arg)
		if len(matched) == 0 {
			return fmt.Errorf("no installed tools match %q\n\nInstalled tools:\n%s", arg, listInstalled(m))
		}
		entries = append(entries, matched...)
	}

	for _, e := range entries {
		owner, name := SplitRepo(e.Repo)
		fmt.Printf("→ Purging %s/%s...\n", owner, name)

		// Remove all shims/symlinks.
		for _, alias := range e.Aliases {
			if err := shim.Remove(alias); err != nil {
				fmt.Printf("  ⚠ Could not remove shim %q: %v\n", alias, err)
			} else {
				fmt.Printf("  Removed command: %s\n", alias)
			}
		}

		// Remove storage directory.
		repoDir := filepath.Join(platform.StorageDir(), "github.com", owner, name)
		if err := os.RemoveAll(repoDir); err != nil {
			fmt.Printf("  ⚠ Could not remove %s: %v\n", repoDir, err)
		} else {
			fmt.Printf("  Removed storage: %s\n", repoDir)
		}

		// Remove from manifest.
		m.Remove(e.Repo)
	}

	if err := manifest.Save(m); err != nil {
		return fmt.Errorf("saving manifest: %w", err)
	}

	fmt.Printf("\n✓ Purged %d tool(s)\n", len(entries))
	return nil
}
