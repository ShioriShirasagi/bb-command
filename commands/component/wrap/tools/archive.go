package tools

import (
	"archive/zip"
	"os"
)

// ZipArchive .zip形式にする
func ZipArchive(basePath string, archivedPath string) error {
	outFile, err := os.Create(archivedPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	zw := zip.NewWriter(outFile)
	defer zw.Close()

	return zw.AddFS(os.DirFS(basePath))
}
