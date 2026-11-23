package tools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Release struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// FindGitHubComponent 条件に合うアセットのダウンロードURLとファイル名をGitHubから検索して返す
func FindGitHubComponent(componentOwner string, componentName string, version string) (downloadURL string, fileName string, err error) {
	var apiURL string
	if version == "" || version == "latest" {
		apiURL = fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", componentOwner, componentName)
	} else {
		apiURL = fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", componentOwner, componentName, version)
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "go-release-downloader")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("APIエラーが発生しました: %s", resp.Status)
	}

	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", "", err
	}

	// 条件に合うアセットを検索
	for _, asset := range rel.Assets {
		if strings.HasSuffix(asset.Name, ".zip") {
			return asset.BrowserDownloadURL, asset.Name, nil
		}
	}

	return "", "", fmt.Errorf("指定されたコンポーネントはありませんでした (version: %s)", rel.TagName)
}
