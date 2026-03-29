package checksum

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/altlimit/alt/internal/github"
)

func TestFindChecksumAsset(t *testing.T) {
	tests := []struct {
		name     string
		assets   []github.Asset
		wantName string
		wantNil  bool
	}{
		{
			name:     "checksums.txt",
			assets:   []github.Asset{{Name: "app.tar.gz"}, {Name: "checksums.txt"}},
			wantName: "checksums.txt",
		},
		{
			name:     "SHA256SUMS",
			assets:   []github.Asset{{Name: "app.tar.gz"}, {Name: "SHA256SUMS.txt"}},
			wantName: "SHA256SUMS.txt",
		},
		{
			name:     ".sha256 extension",
			assets:   []github.Asset{{Name: "app.tar.gz"}, {Name: "app.tar.gz.sha256"}},
			wantName: "app.tar.gz.sha256",
		},
		{
			name:     "contains checksum",
			assets:   []github.Asset{{Name: "app.tar.gz"}, {Name: "my_checksum_file.txt"}},
			wantName: "my_checksum_file.txt",
		},
		{
			name:    "no checksum file",
			assets:  []github.Asset{{Name: "app.tar.gz"}, {Name: "README.md"}},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindChecksumAsset(tt.assets)
			if tt.wantNil {
				if result != nil {
					t.Errorf("expected nil, got %s", result.Name)
				}
				return
			}
			if result == nil {
				t.Fatal("expected a checksum asset, got nil")
			}
			if result.Name != tt.wantName {
				t.Errorf("got %s, want %s", result.Name, tt.wantName)
			}
		})
	}
}

func TestVerifyFile_Correct(t *testing.T) {
	dir := t.TempDir()

	// Create a target file.
	targetPath := filepath.Join(dir, "myfile.bin")
	content := []byte("hello world test content")
	if err := os.WriteFile(targetPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	// Compute its hash.
	h := sha256.Sum256(content)
	hash := hex.EncodeToString(h[:])

	// Create checksum file.
	checksumPath := filepath.Join(dir, "checksums.txt")
	checksumContent := hash + "  myfile.bin\n"
	if err := os.WriteFile(checksumPath, []byte(checksumContent), 0644); err != nil {
		t.Fatal(err)
	}

	if err := VerifyFile(checksumPath, targetPath); err != nil {
		t.Fatalf("expected verification to pass: %v", err)
	}
}

func TestVerifyFile_Mismatch(t *testing.T) {
	dir := t.TempDir()

	targetPath := filepath.Join(dir, "myfile.bin")
	if err := os.WriteFile(targetPath, []byte("actual content"), 0644); err != nil {
		t.Fatal(err)
	}

	checksumPath := filepath.Join(dir, "checksums.txt")
	checksumContent := "0000000000000000000000000000000000000000000000000000000000000000  myfile.bin\n"
	if err := os.WriteFile(checksumPath, []byte(checksumContent), 0644); err != nil {
		t.Fatal(err)
	}

	err := VerifyFile(checksumPath, targetPath)
	if err == nil {
		t.Fatal("expected checksum mismatch error")
	}
}

func TestVerifyFile_FileNotInChecksums(t *testing.T) {
	dir := t.TempDir()

	targetPath := filepath.Join(dir, "myfile.bin")
	if err := os.WriteFile(targetPath, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	checksumPath := filepath.Join(dir, "checksums.txt")
	checksumContent := "abcdef1234567890  otherfile.bin\n"
	if err := os.WriteFile(checksumPath, []byte(checksumContent), 0644); err != nil {
		t.Fatal(err)
	}

	err := VerifyFile(checksumPath, targetPath)
	if err == nil {
		t.Fatal("expected 'not found' error")
	}
}
