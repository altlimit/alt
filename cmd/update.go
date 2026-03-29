package cmd

import (
	"fmt"
	"strings"

	"github.com/altlimit/alt/internal/github"
	"github.com/altlimit/alt/internal/manifest"
)

// Update handles `alt update [repo]`.
// If repo is specified, updates just that tool. Otherwise updates all.
func Update(args []string) error {
	m, err := manifest.Load()
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	if len(m.Entries) == 0 {
		return fmt.Errorf("no tools installed — use 'alt install user/repo' first")
	}

	var entries []manifest.Entry
	if len(args) > 0 {
		repo, _, err := ParseRepoArg(args[0])
		if err != nil {
			return err
		}
		e := m.FindByRepo(repo)
		if e == nil {
			return fmt.Errorf("%s is not installed\n\nInstalled tools:\n%s", repo, listInstalled(m))
		}
		entries = []manifest.Entry{*e}
	} else {
		entries = m.Entries
	}

	client := github.NewClient()
	var updated, upToDate, failed int

	for _, e := range entries {
		owner, name := SplitRepo(e.Repo)
		fmt.Printf("→ Checking %s/%s (current: %s)...", owner, name, e.Version)

		rel, err := client.GetLatestRelease(owner, name)
		if err != nil {
			fmt.Printf(" error: %v\n", err)
			failed++
			continue
		}

		if rel.TagName == e.Version {
			fmt.Println(" up to date")
			upToDate++
			continue
		}

		fmt.Printf(" %s available\n", rel.TagName)

		// Re-install at the new tag.
		if err := Install([]string{e.Repo + "@" + rel.TagName}); err != nil {
			fmt.Printf("  ✗ Update failed: %v\n", err)
			failed++
			continue
		}
		updated++
	}

	fmt.Printf("\nSummary: %d updated, %d up to date", updated, upToDate)
	if failed > 0 {
		fmt.Printf(", %d failed", failed)
	}
	fmt.Println()

	return nil
}

func listInstalled(m *manifest.Manifest) string {
	var lines []string
	for _, e := range m.Entries {
		lines = append(lines, fmt.Sprintf("  %s (%s)", e.Repo, e.Version))
	}
	return strings.Join(lines, "\n")
}
