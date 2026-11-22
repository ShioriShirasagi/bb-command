package tools

import "os"

// FileExists ファイルの存在確認
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
