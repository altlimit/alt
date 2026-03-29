package scoring

import (
	"testing"

	"github.com/altlimit/alt/internal/github"
)

func TestScoreAssets_LinuxAmd64(t *testing.T) {
	assets := []github.Asset{
		{Name: "tool-linux-amd64.tar.gz"},
		{Name: "tool-linux-arm64.tar.gz"},
		{Name: "tool-darwin-amd64.tar.gz"},
		{Name: "tool-darwin-arm64.tar.gz"},
		{Name: "tool-windows-amd64.zip"},
	}

	scored := ScoreAssets(assets, "linux", "amd64")
	if len(scored) == 0 {
		t.Fatal("expected at least one scored asset")
	}

	best := scored[0]
	if best.Asset.Name != "tool-linux-amd64.tar.gz" {
		t.Errorf("expected best asset to be linux-amd64, got %s", best.Asset.Name)
	}
	if best.Score != 220 { // OS(100) + Arch(100) + Archive(20)
		t.Errorf("expected score 220, got %d", best.Score)
	}
	if !best.OS || !best.Arch {
		t.Error("expected OS and Arch to both match")
	}
}

func TestScoreAssets_DarwinArm64(t *testing.T) {
	assets := []github.Asset{
		{Name: "myapp_Darwin_x86_64.tar.gz"},
		{Name: "myapp_Darwin_arm64.tar.gz"},
		{Name: "myapp_Linux_x86_64.tar.gz"},
		{Name: "myapp_Linux_arm64.tar.gz"},
	}

	best := BestAsset(assets, "darwin", "arm64")
	if best == nil {
		t.Fatal("expected a best asset")
	}
	if best.Asset.Name != "myapp_Darwin_arm64.tar.gz" {
		t.Errorf("expected Darwin_arm64, got %s", best.Asset.Name)
	}
}

func TestScoreAssets_PrefersRawBinaryOverArchive(t *testing.T) {
	assets := []github.Asset{
		{Name: "tool-linux-amd64"},         // raw binary: 100+100+50 = 250
		{Name: "tool-linux-amd64.tar.gz"},  // archive: 100+100+20 = 220
	}

	best := BestAsset(assets, "linux", "amd64")
	if best == nil {
		t.Fatal("expected a best asset")
	}
	if best.Asset.Name != "tool-linux-amd64" {
		t.Errorf("expected raw binary, got %s", best.Asset.Name)
	}
	if best.Score != 250 {
		t.Errorf("expected score 250, got %d", best.Score)
	}
}

func TestScoreAssets_PenalizesInstallers(t *testing.T) {
	assets := []github.Asset{
		{Name: "tool-darwin-amd64.pkg"},
		{Name: "tool-darwin-amd64.tar.gz"},
	}

	best := BestAsset(assets, "darwin", "amd64")
	if best == nil {
		t.Fatal("expected a best asset")
	}
	if best.Asset.Name != "tool-darwin-amd64.tar.gz" {
		t.Errorf("expected tar.gz over .pkg, got %s", best.Asset.Name)
	}
}

func TestScoreAssets_SkipsChecksumFiles(t *testing.T) {
	assets := []github.Asset{
		{Name: "checksums.txt"},
		{Name: "tool-linux-amd64.sha256"},
		{Name: "tool-linux-amd64.tar.gz"},
	}

	scored := ScoreAssets(assets, "linux", "amd64")
	for _, s := range scored {
		if s.Asset.Name == "checksums.txt" || s.Asset.Name == "tool-linux-amd64.sha256" {
			t.Errorf("checksum file should have been filtered: %s", s.Asset.Name)
		}
	}
}

func TestScoreAssets_WindowsExePreferred(t *testing.T) {
	assets := []github.Asset{
		{Name: "tool_windows_amd64.exe"},
		{Name: "tool_windows_amd64.zip"},
		{Name: "tool_windows_amd64.msi"},
	}

	best := BestAsset(assets, "windows", "amd64")
	if best == nil {
		t.Fatal("expected a best asset")
	}
	if best.Asset.Name != "tool_windows_amd64.exe" {
		t.Errorf("expected .exe, got %s", best.Asset.Name)
	}
}

func TestScoreAssets_NoMatch(t *testing.T) {
	assets := []github.Asset{
		{Name: "tool-linux-amd64.tar.gz"},
	}

	best := BestAsset(assets, "windows", "arm64")
	if best != nil {
		t.Errorf("expected no match for windows/arm64, got %s (score %d)", best.Asset.Name, best.Score)
	}
}

func TestScoreAssets_MacOSAliases(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{"darwin", "tool-darwin-amd64.tar.gz"},
		{"macos", "tool-macos-amd64.tar.gz"},
		{"osx", "tool-osx-amd64.tar.gz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assets := []github.Asset{{Name: tt.filename}}
			best := BestAsset(assets, "darwin", "amd64")
			if best == nil {
				t.Fatalf("expected %s to match darwin", tt.filename)
			}
			if !best.OS {
				t.Errorf("expected OS match for %s", tt.filename)
			}
		})
	}
}

func TestScoreAssets_ArchAliases(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		goarch   string
	}{
		{"x86_64", "tool-linux-x86_64.tar.gz", "amd64"},
		{"x64", "tool-linux-x64.tar.gz", "amd64"},
		{"aarch64", "tool-linux-aarch64.tar.gz", "arm64"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assets := []github.Asset{{Name: tt.filename}}
			best := BestAsset(assets, "linux", tt.goarch)
			if best == nil {
				t.Fatalf("expected %s to match %s", tt.filename, tt.goarch)
			}
			if !best.Arch {
				t.Errorf("expected arch match for %s → %s", tt.filename, tt.goarch)
			}
		})
	}
}

func TestScoreAssets_WinNotInDarwin(t *testing.T) {
	// Regression: "win" must NOT match inside "darwin".
	assets := []github.Asset{
		{Name: "darwin.tgz"},
		{Name: "win.zip"},
	}

	best := BestAsset(assets, "windows", "amd64")
	if best == nil {
		t.Fatal("expected a match for windows")
	}
	if best.Asset.Name != "win.zip" {
		t.Errorf("expected win.zip, got %s", best.Asset.Name)
	}

	// "darwin.tgz" should NOT match windows.
	scored := ScoreAssets(assets, "windows", "amd64")
	for _, s := range scored {
		if s.Asset.Name == "darwin.tgz" && s.OS {
			t.Error("darwin.tgz should NOT match windows OS")
		}
	}
}

func TestScoreAssets_SitegenWindows(t *testing.T) {
	// Real-world: altlimit/sitegen releases.
	assets := []github.Asset{
		{Name: "darwin-arm64.tgz"},
		{Name: "darwin.tgz"},
		{Name: "linux.tgz"},
		{Name: "win.zip"},
	}

	best := BestAsset(assets, "windows", "amd64")
	if best == nil {
		t.Fatal("expected a match")
	}
	if best.Asset.Name != "win.zip" {
		t.Errorf("expected win.zip for windows, got %s", best.Asset.Name)
	}
}

func TestScoreAssets_AltclawWindows(t *testing.T) {
	// Real-world: altlimit/altclaw releases.
	assets := []github.Asset{
		{Name: "altclaw-cli-darwin-amd64.tgz"},
		{Name: "altclaw-cli-darwin-arm64.tgz"},
		{Name: "altclaw-cli-linux-amd64.tgz"},
		{Name: "altclaw-cli-linux-arm64.tgz"},
		{Name: "altclaw-cli-windows-amd64.zip"},
	}

	best := BestAsset(assets, "windows", "amd64")
	if best == nil {
		t.Fatal("expected a match")
	}
	if best.Asset.Name != "altclaw-cli-windows-amd64.zip" {
		t.Errorf("expected windows asset, got %s", best.Asset.Name)
	}
}

