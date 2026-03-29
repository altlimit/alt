package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/altlimit/alt/internal/platform"
)

// Entry represents a single installed tool.
type Entry struct {
	Repo        string   `json:"repo"`         // "altlimit/sitegen"
	Version     string   `json:"version"`      // "v1.1.0"
	Binary      string   `json:"binary"`       // absolute path to the binary
	Aliases     []string `json:"aliases"`       // ["sitegen", "sg"]
	InstalledAt string   `json:"installed_at"` // RFC3339 timestamp
}

// Manifest holds the state of all installed tools.
type Manifest struct {
	Entries []Entry `json:"entries"`
}

// Load reads the manifest from disk. Returns an empty manifest if the file doesn't exist.
func Load() (*Manifest, error) {
	path := platform.ManifestPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Manifest{}, nil
		}
		return nil, fmt.Errorf("reading manifest at %s: %w", path, err)
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing manifest: %w", err)
	}
	return &m, nil
}

// Save writes the manifest to disk with pretty JSON.
func Save(m *Manifest) error {
	path := platform.ManifestPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating manifest directory: %w", err)
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding manifest: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing manifest: %w", err)
	}
	return nil
}

// FindByRepo returns the entry matching "user/repo", or nil.
func (m *Manifest) FindByRepo(repo string) *Entry {
	repo = strings.ToLower(repo)
	for i := range m.Entries {
		if strings.ToLower(m.Entries[i].Repo) == repo {
			return &m.Entries[i]
		}
	}
	return nil
}

// FindByAlias returns the entry that has the given alias, or nil.
func (m *Manifest) FindByAlias(alias string) *Entry {
	alias = strings.ToLower(alias)
	for i := range m.Entries {
		for _, a := range m.Entries[i].Aliases {
			if strings.ToLower(a) == alias {
				return &m.Entries[i]
			}
		}
	}
	return nil
}

// AddOrUpdate inserts or replaces an entry by repo.
func (m *Manifest) AddOrUpdate(e Entry) {
	if e.InstalledAt == "" {
		e.InstalledAt = time.Now().UTC().Format(time.RFC3339)
	}
	for i := range m.Entries {
		if strings.EqualFold(m.Entries[i].Repo, e.Repo) {
			m.Entries[i] = e
			return
		}
	}
	m.Entries = append(m.Entries, e)
}

// Remove deletes an entry by repo name.
func (m *Manifest) Remove(repo string) {
	repo = strings.ToLower(repo)
	entries := m.Entries[:0]
	for _, e := range m.Entries {
		if strings.ToLower(e.Repo) != repo {
			entries = append(entries, e)
		}
	}
	m.Entries = entries
}

// MatchEntries returns entries whose repo matches the filter.
// Filter can be "user/repo" (exact) or "user" (all repos by that user).
func (m *Manifest) MatchEntries(filter string) []Entry {
	filter = strings.ToLower(filter)
	var matched []Entry

	for _, e := range m.Entries {
		lower := strings.ToLower(e.Repo)
		if lower == filter || strings.HasPrefix(lower, filter+"/") {
			matched = append(matched, e)
		}
	}
	return matched
}

// marshalManifest encodes a manifest to JSON bytes.
func marshalManifest(m *Manifest) ([]byte, error) {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

// loadFromPath reads and parses a manifest from an arbitrary path.
func loadFromPath(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

