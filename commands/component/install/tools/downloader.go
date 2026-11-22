package tools

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// GitHubComponentDownloader GitHubからコンポーネントをダウンロードする
func GitHubPackageDownloader(componentOwner string, componentRepository string) (string, error) {
	componentUrl := "https://github.com/" + componentOwner + "/" + componentRepository + "/releases/latest/download/" + componentRepository + ".zip"

	resp, err := http.Get(componentUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println(fmt.Errorf("ダウンロード失敗: HTTP %d", resp.StatusCode))
		return "", err
	}

	downloadPath := componentRepository + ".zip"
	out, err := os.Create(downloadPath)
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Println(err)
		return "", err
	}

	return downloadPath, nil
}
