package archive

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Extract detects the archive format by filename and extracts it to dest.
// Returns the list of extracted file paths.
func Extract(src, dest string) ([]string, error) {
	lower := strings.ToLower(src)
	switch {
	case strings.HasSuffix(lower, ".tar.gz"), strings.HasSuffix(lower, ".tgz"):
		return extractTarGz(src, dest)
	case strings.HasSuffix(lower, ".zip"):
		return extractZip(src, dest)
	default:
		return nil, fmt.Errorf("unsupported archive format: %s", filepath.Base(src))
	}
}

// IsArchive returns true if the filename looks like a supported archive.
func IsArchive(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, ".tar.gz") ||
		strings.HasSuffix(lower, ".tgz") ||
		strings.HasSuffix(lower, ".zip")
}

func extractTarGz(src, dest string) ([]string, error) {
	f, err := os.Open(src)
	if err != nil {
		return nil, fmt.Errorf("opening archive %s: %w", src, err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("creating gzip reader: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	var extracted []string

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading tar entry: %w", err)
		}

		// Sanitize path to prevent directory traversal.
		target := filepath.Join(dest, filepath.Clean(hdr.Name))
		if !strings.HasPrefix(target, filepath.Clean(dest)+string(os.PathSeparator)) && target != filepath.Clean(dest) {
			continue // skip entries that escape the dest directory
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return nil, fmt.Errorf("creating directory %s: %w", target, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return nil, fmt.Errorf("creating parent directory: %w", err)
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode)|0755)
			if err != nil {
				return nil, fmt.Errorf("creating file %s: %w", target, err)
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return nil, fmt.Errorf("extracting %s: %w", hdr.Name, err)
			}
			out.Close()
			extracted = append(extracted, target)
		}
	}

	return extracted, nil
}

func extractZip(src, dest string) ([]string, error) {
	r, err := zip.OpenReader(src)
	if err != nil {
		return nil, fmt.Errorf("opening zip %s: %w", src, err)
	}
	defer r.Close()

	var extracted []string

	for _, f := range r.File {
		// Sanitize path to prevent directory traversal.
		target := filepath.Join(dest, filepath.Clean(f.Name))
		if !strings.HasPrefix(target, filepath.Clean(dest)+string(os.PathSeparator)) && target != filepath.Clean(dest) {
			continue
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return nil, fmt.Errorf("creating directory %s: %w", target, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return nil, fmt.Errorf("creating parent directory: %w", err)
		}

		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("opening zipentry %s: %w", f.Name, err)
		}

		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode()|0755)
		if err != nil {
			rc.Close()
			return nil, fmt.Errorf("creating file %s: %w", target, err)
		}

		if _, err := io.Copy(out, rc); err != nil {
			out.Close()
			rc.Close()
			return nil, fmt.Errorf("extracting %s: %w", f.Name, err)
		}

		out.Close()
		rc.Close()
		extracted = append(extracted, target)
	}

	return extracted, nil
}
