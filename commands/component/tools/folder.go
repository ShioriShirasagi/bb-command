package tools

import "os"

// RemoveDirectory ディレクトリを削除する
func RemoveDirectory(path string) error {
	return os.RemoveAll(path)
}
