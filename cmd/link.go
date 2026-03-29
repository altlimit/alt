package cmd

import (
	"fmt"

	"github.com/altlimit/alt/internal/manifest"
	"github.com/altlimit/alt/internal/shim"
)

// Link handles `alt link repo[@tag] <alias>`.
func Link(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: alt link <user/repo[@tag]> <alias>\n\nCreates an alias for an installed tool.\n\nExamples:\n  alt link altlimit/sitegen sg\n  alt link altlimit/altclaw ac")
	}

	repo, _, err := ParseRepoArg(args[0])
	if err != nil {
		return err
	}
	alias := args[1]

	if alias == "" {
		return fmt.Errorf("alias cannot be empty")
	}

	m, err := manifest.Load()
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	entry := m.FindByRepo(repo)
	if entry == nil {
		return fmt.Errorf("%s is not installed\n\nInstall it first with: alt install %s", repo, repo)
	}

	// Check if alias is already in use by another tool.
	existing := m.FindByAlias(alias)
	if existing != nil && existing.Repo != entry.Repo {
		return fmt.Errorf("alias %q is already used by %s (%s)\n\nUse a different alias or purge %s first.",
			alias, existing.Repo, existing.Version, existing.Repo)
	}

	// Create the shim/symlink.
	if err := shim.Create(entry.Binary, alias); err != nil {
		return fmt.Errorf("creating shim for alias %q: %w", alias, err)
	}

	// Add alias to entry if not already present.
	hasAlias := false
	for _, a := range entry.Aliases {
		if a == alias {
			hasAlias = true
			break
		}
	}
	if !hasAlias {
		entry.Aliases = append(entry.Aliases, alias)
		if err := manifest.Save(m); err != nil {
			return fmt.Errorf("saving manifest: %w", err)
		}
	}

	fmt.Printf("✓ Linked %s → %s\n", alias, entry.Repo)
	return nil
}
