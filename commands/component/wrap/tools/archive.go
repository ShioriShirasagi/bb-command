package tools

import (
	"archive/zip"
	"os"
)

// TarGzArchive .tar.gz形式にする
func TarGzArchive(basePath string, archivedPath string) error {
	outFile, err := os.Create(archivedPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	zw := zip.NewWriter(outFile)
	defer zw.Close()

	return zw.AddFS(os.DirFS(basePath))
}
