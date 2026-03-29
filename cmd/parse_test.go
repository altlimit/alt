package cmd

import "testing"

func TestParseRepoArg_Valid(t *testing.T) {
	tests := []struct {
		input    string
		wantRepo string
		wantTag  string
	}{
		{"cli/cli", "cli/cli", ""},
		{"junegunn/fzf@v0.50.0", "junegunn/fzf", "v0.50.0"},
		{"altlimit/sitegen@latest", "altlimit/sitegen", "latest"},
		{"user/repo@v1.2.3-beta", "user/repo", "v1.2.3-beta"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			repo, tag, err := ParseRepoArg(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if repo != tt.wantRepo {
				t.Errorf("repo = %q, want %q", repo, tt.wantRepo)
			}
			if tag != tt.wantTag {
				t.Errorf("tag = %q, want %q", tag, tt.wantTag)
			}
		})
	}
}

func TestParseRepoArg_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"no slash", "justarepo"},
		{"empty owner", "/repo"},
		{"empty name", "owner/"},
		{"empty tag", "owner/repo@"},
		{"too many parts", "a/b/c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := ParseRepoArg(tt.input)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestSplitRepo(t *testing.T) {
	owner, name := SplitRepo("altlimit/sitegen")
	if owner != "altlimit" {
		t.Errorf("owner = %q, want %q", owner, "altlimit")
	}
	if name != "sitegen" {
		t.Errorf("name = %q, want %q", name, "sitegen")
	}
}
