package tools

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var client = &http.Client{Timeout: 60 * time.Second}

// DownloadFile downloadURLの内容を指定したpathに保存する
func DownloadFile(downloadURL string, path string) error {
	resp, err := client.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("ダウンロードに失敗: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ダウンロードに失敗: HTTP %d (URL: %s)", resp.StatusCode, downloadURL)
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), "download-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) // Rename成功後はファイルが無いので無害

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		return fmt.Errorf("ファイル書き込みに失敗: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmp.Name(), path)
}
