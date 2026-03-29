package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/altlimit/alt/internal/github"
	"github.com/altlimit/alt/internal/platform"
	"github.com/altlimit/alt/internal/scoring"
	"github.com/altlimit/alt/internal/shim"
)

// SelfUpdate handles `alt self-update`.
// Updates the alt binary itself to the latest release.
func SelfUpdate(currentVersion string) error {
	const repo = "altlimit/alt"
	owner, name := SplitRepo(repo)
	client := github.NewClient()

	fmt.Printf("→ Checking for updates (current: v%s)...", currentVersion)

	rel, err := client.GetLatestRelease(owner, name)
	if err != nil {
		fmt.Println()
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	// Compare versions (strip leading 'v' for comparison).
	latest := rel.TagName
	latestClean := latest
	if len(latestClean) > 0 && latestClean[0] == 'v' {
		latestClean = latestClean[1:]
	}

	if latestClean == currentVersion {
		fmt.Println(" up to date")
		return nil
	}

	fmt.Printf(" %s available\n", latest)

	if len(rel.Assets) == 0 {
		return fmt.Errorf("release %s has no downloadable assets", latest)
	}

	// Find the right binary.
	goos := platform.DetectOS()
	goarch := platform.DetectArch()
	best := scoring.BestAsset(rel.Assets, goos, goarch)
	if best == nil {
		return fmt.Errorf("no compatible binary found for %s/%s", goos, goarch)
	}

	fmt.Printf("  Downloading %s...", best.Asset.Name)

	// Download to a temp file first.
	internalDir := platform.InternalDir()
	if err := platform.EnsureDir(internalDir); err != nil {
		return fmt.Errorf("creating internal directory: %w", err)
	}

	tmpPath := filepath.Join(internalDir, "alt.tmp")
	err = client.DownloadAsset(best.Asset.BrowserDownloadURL, tmpPath, func(written, total int64) {
		if total > 0 {
			pct := float64(written) / float64(total) * 100
			fmt.Printf("\r  Downloading %s... %.0f%%", best.Asset.Name, pct)
		}
	})
	fmt.Println()
	if err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("download failed: %w", err)
	}

	// Determine the final binary name.
	binaryName := "alt"
	if runtime.GOOS == "windows" {
		binaryName = "alt.exe"
	}
	destPath := filepath.Join(internalDir, binaryName)

	// On Windows, can't overwrite a running exe — rename the old one first.
	if runtime.GOOS == "windows" {
		oldPath := destPath + ".old"
		os.Remove(oldPath)
		os.Rename(destPath, oldPath)
	}

	// Move temp to final location.
	if err := os.Rename(tmpPath, destPath); err != nil {
		return fmt.Errorf("replacing binary: %w", err)
	}
	os.Chmod(destPath, 0755)

	// Update the symlink/shim in bin dir.
	if err := shim.Create(destPath, "alt"); err != nil {
		return fmt.Errorf("updating alt shim: %w", err)
	}

	fmt.Printf("\n✓ Updated alt to %s\n", latest)
	return nil
}
