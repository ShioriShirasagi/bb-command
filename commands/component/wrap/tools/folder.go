package tools

import (
	"errors"
	"io"
	"io/fs"
	"os"
)

// FolderGenerator フォルダを作成する
func FolderGenerator(path string) error {
	return os.MkdirAll(path, 0755)
}

// DirectoryExists ディレクトリが存在するかどうか確認する
func DirectoryExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}

// CopyDirectory ディレクトリの内容ごと丸々コピーする
func CopyDirectory(beforePath string, afterPath string) error {
	return os.CopyFS(afterPath, os.DirFS(beforePath))
}

// CopyFile ファイルのをコピーする
func CopyFile(beforePath string, afterPath string) error {
	in, err := os.Open(beforePath)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(afterPath)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := out.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()

}
