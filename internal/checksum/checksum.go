package checksum

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/altlimit/alt/internal/github"
)

// FindChecksumAsset scans the release assets for a checksums/sha256 file.
func FindChecksumAsset(assets []github.Asset) *github.Asset {
	candidates := []string{"checksums.txt", "sha256sums.txt", "sha256sum.txt"}

	for _, a := range assets {
		lower := strings.ToLower(a.Name)

		// Exact match on common names.
		for _, c := range candidates {
			if lower == c {
				return &a
			}
		}

		// Ends with .sha256 or .sha256sum.
		if strings.HasSuffix(lower, ".sha256") || strings.HasSuffix(lower, ".sha256sum") {
			return &a
		}

		// Contains "checksum" in the name.
		if strings.Contains(lower, "checksum") {
			return &a
		}
	}
	return nil
}

// VerifyFile checks that targetFile matches the hash listed in checksumFile.
// The checksum file format is: <hex-hash>  <filename> (or <hex-hash> <filename>).
func VerifyFile(checksumFile, targetFile string) error {
	// Compute SHA256 of the target.
	actual, err := hashFile(targetFile)
	if err != nil {
		return fmt.Errorf("computing hash of %s: %w", targetFile, err)
	}

	// Parse the checksum file.
	expected, err := findHash(checksumFile, targetFile)
	if err != nil {
		return err
	}

	if !strings.EqualFold(actual, expected) {
		return fmt.Errorf("checksum mismatch for %s\n  expected: %s\n  got:      %s", targetFile, expected, actual)
	}

	return nil
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func findHash(checksumFile, targetFile string) (string, error) {
	f, err := os.Open(checksumFile)
	if err != nil {
		return "", fmt.Errorf("opening checksum file: %w", err)
	}
	defer f.Close()

	targetBase := strings.ToLower(filepath.Base(targetFile))
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Format: <hash>  <filename> or <hash> <filename>
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		hash := parts[0]
		name := strings.ToLower(parts[len(parts)-1])
		// Strip leading * (some BSD-style checksum files).
		name = strings.TrimPrefix(name, "*")

		if name == targetBase || strings.HasSuffix(name, "/"+targetBase) {
			return hash, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("reading checksum file: %w", err)
	}

	return "", fmt.Errorf("no checksum entry found for %s in %s", filepath.Base(targetFile), checksumFile)
}
