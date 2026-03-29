package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/altlimit/alt/internal/archive"
	"github.com/altlimit/alt/internal/checksum"
	"github.com/altlimit/alt/internal/github"
	"github.com/altlimit/alt/internal/platform"
	"github.com/altlimit/alt/internal/scoring"
)

// Run handles `alt run user/repo[@tag] [args...]`.
// Downloads and executes a tool without permanently installing it.
func Run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: alt run <user/repo[@tag]> [args...]\n\nRuns a tool without installing it permanently.\n\nExamples:\n  alt run altlimit/sitegen --help\n  alt run altlimit/taskr@v0.1.7 build")
	}

	// Find the repo arg (contains "/"). Everything before it is alt flags,
	// everything after it is tool args.
	repoIdx := -1
	for i, a := range args {
		if strings.Contains(a, "/") {
			repoIdx = i
			break
		}
	}
	if repoIdx == -1 {
		return fmt.Errorf("missing repository argument\n\nUsage: alt run <user/repo[@tag]> [args...]")
	}

	repoArg := args[repoIdx]
	toolArgs := args[repoIdx+1:]

	repo, tag, err := ParseRepoArg(repoArg)
	if err != nil {
		return err
	}

	owner, name := SplitRepo(repo)
	client := github.NewClient()

	// Fetch release.
	fmt.Fprintf(os.Stderr, "→ Fetching %s/%s", owner, name)
	var rel *github.Release
	if tag != "" {
		fmt.Fprintf(os.Stderr, "@%s", tag)
		rel, err = client.GetReleaseByTag(owner, name, tag)
	} else {
		rel, err = client.GetLatestRelease(owner, name)
	}
	fmt.Fprintln(os.Stderr)
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
		return fmt.Errorf("no compatible asset found for %s/%s on %s/%s", owner, name, goos, goarch)
	}

	// Use a cache dir so repeated runs don't re-download.
	cacheDir := filepath.Join(platform.DataDir(), "run", owner, name, tag)
	if err := platform.EnsureDir(cacheDir); err != nil {
		return fmt.Errorf("creating cache directory: %w", err)
	}

	// Check cache first.
	binaryPath := ""
	if cached := findExecutable(listFiles(cacheDir), name); cached != "" {
		binaryPath = cached
	} else {
		// Download.
		downloadPath := filepath.Join(cacheDir, best.Asset.Name)
		fmt.Fprintf(os.Stderr, "  Downloading %s...", best.Asset.Name)
		err = client.DownloadAsset(best.Asset.BrowserDownloadURL, downloadPath, func(written, total int64) {
			if total > 0 {
				pct := float64(written) / float64(total) * 100
				fmt.Fprintf(os.Stderr, "\r  Downloading %s... %.0f%%", best.Asset.Name, pct)
			}
		})
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return fmt.Errorf("download failed: %w", err)
		}

		// Checksum verification.
		checksumAsset := checksum.FindChecksumAsset(rel.Assets)
		if checksumAsset != nil {
			checksumPath := filepath.Join(cacheDir, checksumAsset.Name)
			if dlErr := client.DownloadAsset(checksumAsset.BrowserDownloadURL, checksumPath, nil); dlErr == nil {
				if verifyErr := checksum.VerifyFile(checksumPath, downloadPath); verifyErr != nil {
					os.Remove(downloadPath)
					return fmt.Errorf("checksum verification failed: %w", verifyErr)
				}
			}
		}

		// Extract if archive.
		if archive.IsArchive(downloadPath) {
			extracted, err := archive.Extract(downloadPath, cacheDir)
			if err != nil {
				return fmt.Errorf("extraction failed: %w", err)
			}
			binaryPath = findExecutable(extracted, name)
			if binaryPath == "" {
				return fmt.Errorf("could not find an executable in the extracted archive\n\nExtracted files:\n%s",
					strings.Join(extracted, "\n"))
			}
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

	// Execute the binary.
	fmt.Fprintf(os.Stderr, "  Running %s %s\n\n", filepath.Base(binaryPath), tag)

	cmd := exec.Command(binaryPath, toolArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("executing %s: %w", filepath.Base(binaryPath), err)
	}
	return nil
}
