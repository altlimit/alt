package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/altlimit/alt/internal/manifest"
	"github.com/altlimit/alt/internal/platform"
)

// Versions handles `alt versions <user/repo>`.
// Lists all cached versions of a tool on the local disk.
func Versions(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: alt versions <user/repo>\n\nExample:\n  alt versions altlimit/sitegen")
	}

	repo, _, err := ParseRepoArg(args[0])
	if err != nil {
		return err
	}
	owner, name := SplitRepo(repo)

	repoDir := filepath.Join(platform.StorageDir(), "github.com", owner, name)
	dirEntries, err := os.ReadDir(repoDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s/%s is not installed\n\nInstall it with: alt install %s/%s", owner, name, owner, name)
		}
		return fmt.Errorf("reading storage directory: %w", err)
	}

	// Find active version from manifest.
	m, err := manifest.Load()
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}
	entry := m.FindByRepo(repo)
	activeVersion := ""
	if entry != nil {
		activeVersion = entry.Version
	}

	fmt.Printf("Versions of %s/%s:\n\n", owner, name)
	count := 0
	for _, de := range dirEntries {
		if !de.IsDir() {
			continue
		}
		marker := "  "
		if de.Name() == activeVersion {
			marker = "* "
		}
		fmt.Printf("  %s%s\n", marker, de.Name())
		count++
	}

	if count == 0 {
		fmt.Println("  (no versions found)")
	} else {
		fmt.Printf("\n  (* = active)\n")
	}

	return nil
}
