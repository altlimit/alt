package shim

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/altlimit/alt/internal/platform"
)

// Create creates a symlink (Unix) or .bat shim (Windows) in BinDir
// that points to binaryPath.
func Create(binaryPath, alias string) error {
	binDir := platform.BinDir()
	if err := platform.EnsureDir(binDir); err != nil {
		return fmt.Errorf("creating bin directory %s: %w", binDir, err)
	}

	if runtime.GOOS == "windows" {
		return createBatShim(binaryPath, alias, binDir)
	}
	return createSymlink(binaryPath, alias, binDir)
}

// Remove removes the shim/symlink for the given alias from BinDir.
func Remove(alias string) error {
	binDir := platform.BinDir()

	if runtime.GOOS == "windows" {
		batPath := filepath.Join(binDir, alias+".bat")
		if err := os.Remove(batPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("removing shim %s: %w", batPath, err)
		}
		return nil
	}

	linkPath := filepath.Join(binDir, alias)
	if err := os.Remove(linkPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing symlink %s: %w", linkPath, err)
	}
	return nil
}

func createSymlink(binaryPath, alias, binDir string) error {
	linkPath := filepath.Join(binDir, alias)

	// Remove existing link if present.
	if _, err := os.Lstat(linkPath); err == nil {
		if err := os.Remove(linkPath); err != nil {
			return fmt.Errorf("removing existing symlink %s: %w", linkPath, err)
		}
	}

	if err := os.Symlink(binaryPath, linkPath); err != nil {
		return fmt.Errorf("creating symlink %s -> %s: %w", linkPath, binaryPath, err)
	}
	return nil
}

func createBatShim(binaryPath, alias, binDir string) error {
	batPath := filepath.Join(binDir, alias+".bat")
	content := fmt.Sprintf("@echo off\r\n\"%s\" %%*\r\n", binaryPath)

	if err := os.WriteFile(batPath, []byte(content), 0755); err != nil {
		return fmt.Errorf("writing shim %s: %w", batPath, err)
	}
	return nil
}
