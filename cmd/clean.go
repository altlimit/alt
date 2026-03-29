package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/altlimit/alt/internal/manifest"
	"github.com/altlimit/alt/internal/platform"
)

// Clean handles `alt clean [repo|user]`.
// Removes old version folders but keeps the active version.
func Clean(args []string) error {
	m, err := manifest.Load()
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	var entries []manifest.Entry
	if len(args) > 0 {
		entries = m.MatchEntries(args[0])
	} else {
		entries = m.Entries
	}

	var totalCleaned int
	var totalBytes int64

	for _, e := range entries {
		owner, name := SplitRepo(e.Repo)
		repoDir := filepath.Join(platform.StorageDir(), "github.com", owner, name)

		dirEntries, err := os.ReadDir(repoDir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			fmt.Printf("  ⚠ Could not read %s: %v\n", repoDir, err)
			continue
		}

		for _, de := range dirEntries {
			if !de.IsDir() {
				continue
			}
			if de.Name() == e.Version {
				continue // Keep the active version.
			}

			versionDir := filepath.Join(repoDir, de.Name())
			size := dirSize(versionDir)
			if err := os.RemoveAll(versionDir); err != nil {
				fmt.Printf("  ⚠ Could not remove %s: %v\n", versionDir, err)
				continue
			}

			fmt.Printf("  Removed %s/%s %s\n", owner, name, de.Name())
			totalCleaned++
			totalBytes += size
		}
	}

	// Also clean the run cache.
	runDir := filepath.Join(platform.DataDir(), "run")
	if info, err := os.Stat(runDir); err == nil && info.IsDir() {
		size := dirSize(runDir)
		if size > 0 {
			if err := os.RemoveAll(runDir); err != nil {
				fmt.Printf("  ⚠ Could not remove run cache: %v\n", err)
			} else {
				fmt.Println("  Removed run cache")
				totalCleaned++
				totalBytes += size
			}
		}
	}

	if totalCleaned == 0 {
		fmt.Println("✓ Nothing to clean")
	} else {
		fmt.Printf("\n✓ Cleaned %d item(s), freed %s\n", totalCleaned, formatBytes(totalBytes))
	}
	return nil
}

func dirSize(path string) int64 {
	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

func formatBytes(b int64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case b >= gb:
		return fmt.Sprintf("%.1f GB", float64(b)/float64(gb))
	case b >= mb:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(mb))
	case b >= kb:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(kb))
	default:
		return fmt.Sprintf("%d bytes", b)
	}
}

// listInstalledForClean is a helper to show installed tools (used in error messages).
func listInstalledForClean(m *manifest.Manifest) string {
	var lines []string
	for _, e := range m.Entries {
		lines = append(lines, fmt.Sprintf("  %s (%s)", e.Repo, e.Version))
	}
	return strings.Join(lines, "\n")
}
