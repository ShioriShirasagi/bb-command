package tools

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Zipファイルを解凍する
func ZipExtract(archivedPath, extractedPath string) error {
	r, err := zip.OpenReader(archivedPath)
	if err != nil {
		return err
	}
	defer r.Close()

	if err := os.MkdirAll(extractedPath, 0o755); err != nil {
		return err
	}
	root, err := os.OpenRoot(extractedPath)
	if err != nil {
		return err
	}
	defer root.Close()

	for _, f := range r.File {
		relPath := strings.TrimPrefix(filepath.Clean(f.Name), "./")
		if relPath == "." || relPath == "" {
			continue
		}

		if f.FileInfo().IsDir() {
			if err := root.MkdirAll(relPath, f.Mode()); err != nil {
				return err
			}
			continue
		}

		if dir := filepath.Dir(relPath); dir != "." {
			if err := root.MkdirAll(dir, 0o755); err != nil {
				return err
			}
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		out, err := root.OpenFile(relPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}

		_, copyErr := io.Copy(out, rc)
		closeErr := out.Close()
		rc.Close()

		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}
