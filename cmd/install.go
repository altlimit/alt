package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/altlimit/alt/internal/archive"
	"github.com/altlimit/alt/internal/checksum"
	"github.com/altlimit/alt/internal/github"
	"github.com/altlimit/alt/internal/manifest"
	"github.com/altlimit/alt/internal/platform"
	"github.com/altlimit/alt/internal/scoring"
	"github.com/altlimit/alt/internal/shim"
)

// Install handles `alt install [-f] user/repo[@tag] [user/repo[@tag]...]`.
func Install(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: alt install [-f] <user/repo[@tag]> [...]\n\nOptions:\n  -f, --force    Force re-download even if cached\n\nExamples:\n  alt install altlimit/sitegen\n  alt install altlimit/sitegen altlimit/taskr\n  alt install -f altlimit/sitegen@latest")
	}

	// Parse flags.
	force := false
	var positional []string
	for _, a := range args {
		switch a {
		case "-f", "--force":
			force = true
		default:
			positional = append(positional, a)
		}
	}

	if len(positional) == 0 {
		return fmt.Errorf("missing repository argument\n\nUsage: alt install [-f] <user/repo[@tag]> [...]")
	}

	for i, p := range positional {
		if i > 0 {
			fmt.Println()
		}
		if err := installOne(p, force); err != nil {
			return err
		}
	}
	return nil
}

func installOne(arg string, force bool) error {
	repo, tag, err := ParseRepoArg(arg)
	if err != nil {
		return err
	}

	owner, name := SplitRepo(repo)
	client := github.NewClient()

	// Fetch release.
	fmt.Printf("→ Fetching release for %s/%s", owner, name)
	var rel *github.Release
	if tag != "" {
		fmt.Printf("@%s", tag)
		rel, err = client.GetReleaseByTag(owner, name, tag)
	} else {
		rel, err = client.GetLatestRelease(owner, name)
	}
	fmt.Println()
	if err != nil {
		return fmt.Errorf("failed to fetch release: %w", err)
	}

	if len(rel.Assets) == 0 {
		return fmt.Errorf("release %s has no downloadable assets", rel.TagName)
	}

	tag = rel.TagName

	// Score assets.
	goos := platform.DetectOS()
	goarch := platform.DetectArch()
	best := scoring.BestAsset(rel.Assets, goos, goarch)
	if best == nil {
		return fmt.Errorf("no compatible asset found for %s/%s on %s/%s\n\nAvailable assets:\n%s",
			owner, name, goos, goarch, listAssetNames(rel.Assets))
	}

	fmt.Printf("  Selected: %s (score: %d)\n", best.Asset.Name, best.Score)

	// Prepare storage directory.
	versionDir := filepath.Join(platform.StorageDir(), "github.com", owner, name, tag)
	if err := platform.EnsureDir(versionDir); err != nil {
		return fmt.Errorf("creating storage directory: %w", err)
	}

	// Check if this version is already cached.
	var binaryPath string
	if !force {
		if cached := findExecutable(listFiles(versionDir), name); cached != "" {
			binaryPath = cached
			fmt.Printf("  Cached: %s\n", filepath.Base(binaryPath))
		}
	}

	if binaryPath == "" {
		// Download asset.
		downloadPath := filepath.Join(versionDir, best.Asset.Name)
		fmt.Printf("  Downloading %s...", best.Asset.Name)
		err = client.DownloadAsset(best.Asset.BrowserDownloadURL, downloadPath, func(written, total int64) {
			if total > 0 {
				pct := float64(written) / float64(total) * 100
				fmt.Printf("\r  Downloading %s... %.0f%%", best.Asset.Name, pct)
			}
		})
		fmt.Println()
		if err != nil {
			return fmt.Errorf("download failed: %w", err)
		}

		// Checksum verification.
		checksumAsset := checksum.FindChecksumAsset(rel.Assets)
		if checksumAsset != nil {
			fmt.Printf("  Verifying checksum...")
			checksumPath := filepath.Join(versionDir, checksumAsset.Name)
			if dlErr := client.DownloadAsset(checksumAsset.BrowserDownloadURL, checksumPath, nil); dlErr == nil {
				if verifyErr := checksum.VerifyFile(checksumPath, downloadPath); verifyErr != nil {
					os.Remove(downloadPath)
					return fmt.Errorf("checksum verification failed: %w", verifyErr)
				}
				fmt.Println(" ✓")
			} else {
				fmt.Println(" (skipped, could not download checksum file)")
			}
		}

		// Extract if archive.
		if archive.IsArchive(downloadPath) {
			fmt.Printf("  Extracting...")
			extracted, err := archive.Extract(downloadPath, versionDir)
			if err != nil {
				return fmt.Errorf("extraction failed: %w", err)
			}
			fmt.Printf(" %d files\n", len(extracted))

			binaryPath = findExecutable(extracted, name)
			if binaryPath == "" {
				return fmt.Errorf("could not find an executable in the extracted archive\n\nExtracted files:\n%s",
					strings.Join(extracted, "\n"))
			}

			// Remove the archive to save space.
			os.Remove(downloadPath)
		} else {
			binaryPath = downloadPath
			os.Chmod(binaryPath, 0755)
		}
	}

	// On Windows, ensure the binary has a .exe extension.
	if runtime.GOOS == "windows" && !strings.HasSuffix(strings.ToLower(binaryPath), ".exe") {
		exePath := binaryPath + ".exe"
		if err := os.Rename(binaryPath, exePath); err != nil {
			return fmt.Errorf("renaming binary to .exe: %w", err)
		}
		binaryPath = exePath
	}

	fmt.Printf("  Binary: %s\n", filepath.Base(binaryPath))

	// Default alias is the repo name.
	alias := name

	// Create shim/symlink.
	if err := shim.Create(binaryPath, alias); err != nil {
		return fmt.Errorf("creating shim: %w", err)
	}

	// Update manifest.
	m, err := manifest.Load()
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	existing := m.FindByRepo(repo)
	aliases := []string{alias}
	if existing != nil {
		// Preserve any previously-added aliases.
		for _, a := range existing.Aliases {
			if !strings.EqualFold(a, alias) {
				aliases = append(aliases, a)
			}
		}
	}

	m.AddOrUpdate(manifest.Entry{
		Repo:    repo,
		Version: tag,
		Binary:  binaryPath,
		Aliases: aliases,
	})

	if err := manifest.Save(m); err != nil {
		return fmt.Errorf("saving manifest: %w", err)
	}

	fmt.Printf("\n✓ Installed %s/%s %s → %s\n", owner, name, tag, alias)
	return nil
}

// findExecutable finds the best executable candidate from a list of extracted files.
func findExecutable(files []string, repoName string) string {
	// Prefer a file whose name matches the repo name.
	for _, f := range files {
		base := strings.ToLower(filepath.Base(f))
		baseName := strings.TrimSuffix(base, ".exe")
		if baseName == strings.ToLower(repoName) {
			return f
		}
	}

	// Fall back: any file that is executable (no common non-binary extension).
	nonBinaryExts := map[string]bool{
		".md": true, ".txt": true, ".html": true, ".css": true, ".js": true,
		".json": true, ".yaml": true, ".yml": true, ".toml": true,
		".sh": true, ".bat": true, ".ps1": true, ".1": true,
		".tar.gz": true, ".zip": true, ".tgz": true,
		".license": true, ".licence": true,
		".sha256": true, ".sha512": true, ".sig": true, ".asc": true,
	}

	for _, f := range files {
		ext := strings.ToLower(filepath.Ext(f))
		if nonBinaryExts[ext] {
			continue
		}
		info, err := os.Stat(f)
		if err != nil {
			continue
		}
		if info.IsDir() {
			continue
		}
		// On Unix, check executable bit; on Windows, prefer .exe.
		if ext == ".exe" || info.Mode()&0111 != 0 {
			return f
		}
	}

	// Last resort: first regular file.
	for _, f := range files {
		info, err := os.Stat(f)
		if err == nil && !info.IsDir() {
			return f
		}
	}
	return ""
}

func listAssetNames(assets []github.Asset) string {
	var names []string
	for _, a := range assets {
		names = append(names, "  - "+a.Name)
	}
	return strings.Join(names, "\n")
}

// listFiles returns all file paths (recursively) in a directory.
func listFiles(dir string) []string {
	var files []string
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files
}
