package platform

import (
	"os"
	"path/filepath"
	"runtime"
)

// DetectOS returns the normalized operating system name.
func DetectOS() string {
	switch runtime.GOOS {
	case "darwin":
		return "darwin"
	case "windows":
		return "windows"
	default:
		return "linux"
	}
}

// DetectArch returns the normalized architecture name.
func DetectArch() string {
	switch runtime.GOARCH {
	case "arm64":
		return "arm64"
	case "386":
		return "386"
	default:
		return "amd64"
	}
}

// HomeDir returns the user's home directory or panics.
func HomeDir() string {
	h, err := os.UserHomeDir()
	if err != nil {
		panic("alt: unable to determine home directory: " + err.Error())
	}
	return h
}

// DataDir returns the base data directory for alt.
//
//	Unix:    ~/.local/share/alt/
//	Windows: %LOCALAPPDATA%\alt\
func DataDir() string {
	if runtime.GOOS == "windows" {
		base := os.Getenv("LOCALAPPDATA")
		if base == "" {
			base = filepath.Join(HomeDir(), "AppData", "Local")
		}
		return filepath.Join(base, "alt")
	}
	return filepath.Join(HomeDir(), ".local", "share", "alt")
}

// StorageDir returns the directory where versioned binaries are kept.
//
//	~/.local/share/alt/storage/
func StorageDir() string {
	return filepath.Join(DataDir(), "storage")
}

// BinDir returns the directory where shims/symlinks live.
//
//	Unix:    ~/.local/share/alt/bin/
//	Windows: %LOCALAPPDATA%\alt\bin\
func BinDir() string {
	return filepath.Join(DataDir(), "bin")
}

// ManifestPath returns the path to the manifest JSON file.
func ManifestPath() string {
	return filepath.Join(DataDir(), "manifest.json")
}

// InternalDir returns the directory where the alt binary itself resides.
func InternalDir() string {
	return filepath.Join(DataDir(), "internal")
}

// EnsureDir creates a directory and all parents if they don't exist.
func EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}
