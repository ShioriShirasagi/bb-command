package tools

import (
	"fmt"
	"os"
	"path/filepath"

	commonpkgtools "github.com/ShioriShirasagi/bb-command/commands/component/tools"
)

// componentPaths コンポーネントの配置先を返す
func componentPaths(key string, def commonpkgtools.ComponentDefineJSON) []string {
	var paths []string
	if def.Status.HasFrontend {
		paths = append(paths, frontendComponentsRoot+"/"+key)
	}
	if def.Status.HasBackend {
		paths = append(paths, backendComponentsRoot+"/"+key)
	}
	return paths
}

// UninstallComponent インストールしたコンポーネントを削除する
func UninstallComponent(key string, def commonpkgtools.ComponentDefineJSON) {
	for _, path := range componentPaths(key, def) {
		if err := commonpkgtools.RemoveDirectory(path); err != nil {
			fmt.Println(key + "の削除に失敗しました。以下コマンドで削除してください。")
			fmt.Println("rm -rf " + path)
			continue
		}
		os.Remove(filepath.Dir(path)) // ownerディレクトリが空になっていれば消す
	}
}

// InstallComponent コンポーネントを「owner/name」の配置先にコピーする
func InstallComponent(key string, def commonpkgtools.ComponentDefineJSON, sourcePath string) error {
	for _, path := range componentPaths(key, def) {
		exists, err := commonpkgtools.DirectoryExists(path)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("%s は既に存在します。同名のコンポーネントをアンインストールしてから再度インストールしてください", path)
		}
	}

	if def.Status.HasFrontend {
		err := commonpkgtools.CopyDirectory(sourcePath+"/frontend", frontendComponentsRoot+"/"+key)
		if err != nil {
			UninstallComponent(key, def)
			return err
		}
	}
	if def.Status.HasBackend {
		err := commonpkgtools.CopyDirectory(sourcePath+"/backend", backendComponentsRoot+"/"+key)
		if err != nil {
			UninstallComponent(key, def)
			return err
		}
	}
	return nil
}
