package cmd

import (
	"fmt"
	"strings"

	"github.com/altlimit/alt/internal/manifest"
)

// List handles `alt list [user]`.
// Shows all installed tools, optionally filtered by user/org.
func List(args []string) error {
	m, err := manifest.Load()
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	if len(m.Entries) == 0 {
		fmt.Println("No tools installed.")
		fmt.Println("\nGet started: alt install user/repo")
		return nil
	}

	var entries []manifest.Entry
	if len(args) > 0 {
		entries = m.MatchEntries(args[0])
		if len(entries) == 0 {
			return fmt.Errorf("no installed tools match %q\n\nAll installed tools:\n%s", args[0], listInstalled(m))
		}
		fmt.Printf("Installed tools matching %q:\n\n", args[0])
	} else {
		entries = m.Entries
		fmt.Println("Installed tools:")
	}

	for _, e := range entries {
		aliases := strings.Join(e.Aliases, ", ")
		fmt.Printf("  %-30s %s\n", e.Repo, e.Version)
		if aliases != "" {
			fmt.Printf("  %-30s commands: %s\n", "", aliases)
		}
	}
	fmt.Printf("\n%d tool(s) installed\n", len(entries))

	return nil
}
