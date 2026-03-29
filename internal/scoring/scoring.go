package scoring

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/altlimit/alt/internal/github"
)

// ScoredAsset pairs a GitHub asset with its compatibility score.
type ScoredAsset struct {
	Asset github.Asset
	Score int
	OS    bool
	Arch  bool
}

// osAliases maps our canonical OS names to possible filename substrings.
var osAliases = map[string][]string{
	"darwin":  {"darwin", "macos", "osx", "apple"},
	"linux":   {"linux"},
	"windows": {"windows", "win"},
}

// archAliases maps our canonical Arch names to possible filename substrings.
var archAliases = map[string][]string{
	"amd64": {"amd64", "x86_64", "x64"},
	"arm64": {"arm64", "aarch64"},
	"386":   {"386", "i386", "i686", "x86"},
}

// skipExtensions are file types we never want to select.
var skipExtensions = []string{".sha256", ".sha512", ".md5", ".sig", ".asc", ".pem", ".sbom"}

// ScoreAssets ranks all assets by compatibility and returns them sorted
// from best (highest score) to worst. Assets with score <= 0 are excluded.
func ScoreAssets(assets []github.Asset, goos, goarch string) []ScoredAsset {
	var scored []ScoredAsset

	for _, a := range assets {
		name := strings.ToLower(a.Name)

		// Skip checksum and signature files entirely.
		if isChecksumFile(name) || isSkipExtension(name) {
			continue
		}

		osMatch := matchOS(name, goos)
		archMatch := matchArch(name, goarch)

		// Calculate S = (O × 100) + (A × 100) + P
		s := 0
		if osMatch {
			s += 100
		}
		if archMatch {
			s += 100
		}
		s += preferenceScore(name)

		scored = append(scored, ScoredAsset{
			Asset: a,
			Score: s,
			OS:    osMatch,
			Arch:  archMatch,
		})
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	// Filter out assets that don't match at least OS or Arch.
	var result []ScoredAsset
	for _, s := range scored {
		if s.OS || s.Arch {
			result = append(result, s)
		}
	}
	return result
}

// BestAsset returns the single best-matching asset, or nil if none match.
func BestAsset(assets []github.Asset, goos, goarch string) *ScoredAsset {
	scored := ScoreAssets(assets, goos, goarch)
	if len(scored) == 0 {
		return nil
	}
	return &scored[0]
}

func matchOS(name, goos string) bool {
	aliases, ok := osAliases[goos]
	if !ok {
		return false
	}
	for _, alias := range aliases {
		if containsWord(name, alias) {
			return true
		}
	}
	return false
}

func matchArch(name, goarch string) bool {
	aliases, ok := archAliases[goarch]
	if !ok {
		return false
	}
	for _, alias := range aliases {
		if containsWord(name, alias) {
			return true
		}
	}
	return false
}

// containsWord checks if word appears in s as a standalone segment,
// bounded by start/end of string or common delimiters (-_.).
// This prevents "win" from matching inside "darwin".
func containsWord(s, word string) bool {
	idx := 0
	for {
		pos := strings.Index(s[idx:], word)
		if pos == -1 {
			return false
		}
		pos += idx

		// Check left boundary.
		leftOk := pos == 0 || isBoundary(s[pos-1])
		// Check right boundary.
		end := pos + len(word)
		rightOk := end == len(s) || isBoundary(s[end])

		if leftOk && rightOk {
			return true
		}
		idx = pos + 1
	}
}

func isBoundary(c byte) bool {
	return c == '-' || c == '_' || c == '.' || c == '/' || c == ' '
}

func preferenceScore(name string) int {
	ext := strings.ToLower(filepath.Ext(name))
	// Handle .tar.gz / .tgz
	if strings.HasSuffix(name, ".tar.gz") || ext == ".tgz" {
		return 20
	}
	switch ext {
	case ".zip":
		return 20
	case ".msi", ".pkg", ".deb", ".rpm", ".dmg":
		return -50
	case ".exe":
		return 50
	case "":
		// No extension — likely a raw binary.
		return 50
	default:
		return 0
	}
}

func isChecksumFile(name string) bool {
	lower := strings.ToLower(name)
	if strings.Contains(lower, "checksum") || strings.Contains(lower, "sha256sum") || strings.Contains(lower, "sha512sum") {
		return true
	}
	return false
}

func isSkipExtension(name string) bool {
	lower := strings.ToLower(name)
	for _, ext := range skipExtensions {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}
