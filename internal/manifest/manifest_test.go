package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestManifest(t *testing.T) (string, func()) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")
	orig := os.Getenv("ALT_MANIFEST_PATH")
	// We'll use a helper to override the path for testing.
	return path, func() {
		os.Setenv("ALT_MANIFEST_PATH", orig)
	}
}

func TestManifest_AddAndFind(t *testing.T) {
	m := &Manifest{}

	m.AddOrUpdate(Entry{
		Repo:    "cli/cli",
		Version: "v2.50.0",
		Binary:  "/some/path/gh",
		Aliases: []string{"gh"},
	})

	m.AddOrUpdate(Entry{
		Repo:    "junegunn/fzf",
		Version: "v0.50.0",
		Binary:  "/some/path/fzf",
		Aliases: []string{"fzf"},
	})

	if len(m.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(m.Entries))
	}

	// FindByRepo
	e := m.FindByRepo("cli/cli")
	if e == nil {
		t.Fatal("expected to find cli/cli")
	}
	if e.Version != "v2.50.0" {
		t.Errorf("version = %q, want v2.50.0", e.Version)
	}

	// FindByRepo case insensitive
	e = m.FindByRepo("CLI/CLI")
	if e == nil {
		t.Fatal("expected case-insensitive match for CLI/CLI")
	}

	// FindByAlias
	e = m.FindByAlias("fzf")
	if e == nil {
		t.Fatal("expected to find alias fzf")
	}
	if e.Repo != "junegunn/fzf" {
		t.Errorf("repo = %q, want junegunn/fzf", e.Repo)
	}

	// FindByAlias — not found
	e = m.FindByAlias("nonexistent")
	if e != nil {
		t.Error("expected nil for nonexistent alias")
	}
}

func TestManifest_Update(t *testing.T) {
	m := &Manifest{}

	m.AddOrUpdate(Entry{
		Repo:    "cli/cli",
		Version: "v2.50.0",
		Binary:  "/path/old",
		Aliases: []string{"gh"},
	})

	m.AddOrUpdate(Entry{
		Repo:    "cli/cli",
		Version: "v2.51.0",
		Binary:  "/path/new",
		Aliases: []string{"gh"},
	})

	if len(m.Entries) != 1 {
		t.Fatalf("expected 1 entry after update, got %d", len(m.Entries))
	}
	if m.Entries[0].Version != "v2.51.0" {
		t.Errorf("version = %q, want v2.51.0", m.Entries[0].Version)
	}
	if m.Entries[0].Binary != "/path/new" {
		t.Errorf("binary = %q, want /path/new", m.Entries[0].Binary)
	}
}

func TestManifest_Remove(t *testing.T) {
	m := &Manifest{}

	m.AddOrUpdate(Entry{Repo: "a/b", Version: "v1"})
	m.AddOrUpdate(Entry{Repo: "c/d", Version: "v2"})
	m.AddOrUpdate(Entry{Repo: "e/f", Version: "v3"})

	m.Remove("c/d")

	if len(m.Entries) != 2 {
		t.Fatalf("expected 2 entries after remove, got %d", len(m.Entries))
	}
	if m.FindByRepo("c/d") != nil {
		t.Error("c/d should have been removed")
	}
	if m.FindByRepo("a/b") == nil {
		t.Error("a/b should still exist")
	}
	if m.FindByRepo("e/f") == nil {
		t.Error("e/f should still exist")
	}
}

func TestManifest_MatchEntries(t *testing.T) {
	m := &Manifest{}

	m.AddOrUpdate(Entry{Repo: "altlimit/sitegen", Version: "v1"})
	m.AddOrUpdate(Entry{Repo: "altlimit/alt", Version: "v2"})
	m.AddOrUpdate(Entry{Repo: "junegunn/fzf", Version: "v3"})

	// Match by user
	matches := m.MatchEntries("altlimit")
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches for altlimit, got %d", len(matches))
	}

	// Match by exact repo
	matches = m.MatchEntries("junegunn/fzf")
	if len(matches) != 1 {
		t.Fatalf("expected 1 match for junegunn/fzf, got %d", len(matches))
	}

	// No match
	matches = m.MatchEntries("nobody")
	if len(matches) != 0 {
		t.Fatalf("expected 0 matches for nobody, got %d", len(matches))
	}
}

func TestManifest_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")

	m := &Manifest{}
	m.AddOrUpdate(Entry{
		Repo:    "cli/cli",
		Version: "v2.50.0",
		Binary:  "/path/to/gh",
		Aliases: []string{"gh", "github"},
	})

	// Save
	data, err := marshalManifest(m)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Load back
	loaded, err := loadFromPath(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if len(loaded.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(loaded.Entries))
	}
	e := loaded.Entries[0]
	if e.Repo != "cli/cli" || e.Version != "v2.50.0" {
		t.Errorf("unexpected entry: %+v", e)
	}
	if len(e.Aliases) != 2 || e.Aliases[0] != "gh" || e.Aliases[1] != "github" {
		t.Errorf("unexpected aliases: %v", e.Aliases)
	}
}
