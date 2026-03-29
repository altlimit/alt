package cmd

import (
	"fmt"

	"github.com/altlimit/alt/internal/manifest"
)

// Which handles `alt which <alias>`.
// Displays the storage path of the binary linked to an alias.
func Which(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: alt which <command>\n\nDisplays the storage path of the binary linked to a command.\n\nExamples:\n  alt which fzf\n  alt which sg")
	}

	alias := args[0]

	m, err := manifest.Load()
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	entry := m.FindByAlias(alias)
	if entry == nil {
		// Also try by repo name.
		entry = m.FindByRepo(alias)
	}
	if entry == nil {
		return fmt.Errorf("command %q not found\n\nNo installed tool provides this command.\nUse 'alt install user/repo' to install a tool.", alias)
	}

	fmt.Printf("%s\n", entry.Binary)
	return nil
}
