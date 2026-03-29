package cmd

import (
	"fmt"

	"github.com/altlimit/alt/internal/manifest"
	"github.com/altlimit/alt/internal/shim"
)

// Unlink handles `alt unlink <alias>`.
// Removes a single alias without purging the tool.
func Unlink(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: alt unlink <alias>\n\nRemoves a command alias without uninstalling the tool.\n\nExamples:\n  alt unlink sg\n  alt unlink sitegen")
	}

	alias := args[0]

	m, err := manifest.Load()
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	entry := m.FindByAlias(alias)
	if entry == nil {
		return fmt.Errorf("alias %q not found\n\nNo installed tool provides this command.", alias)
	}

	// Don't allow removing the last alias.
	if len(entry.Aliases) <= 1 {
		return fmt.Errorf("cannot unlink %q — it's the only alias for %s\n\nUse 'alt purge %s' to remove the tool entirely.",
			alias, entry.Repo, entry.Repo)
	}

	// Remove the shim/symlink.
	if err := shim.Remove(alias); err != nil {
		return fmt.Errorf("removing shim for %q: %w", alias, err)
	}

	// Remove alias from manifest entry.
	var remaining []string
	for _, a := range entry.Aliases {
		if a != alias {
			remaining = append(remaining, a)
		}
	}
	entry.Aliases = remaining

	if err := manifest.Save(m); err != nil {
		return fmt.Errorf("saving manifest: %w", err)
	}

	fmt.Printf("✓ Unlinked %s (tool %s still installed as: %s)\n", alias, entry.Repo, remaining[0])
	return nil
}
