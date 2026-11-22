package tools

import (
	"fmt"
	"log"

	commonpkgtools "github.com/ShioriShirasagi/bb-command/commands/component/tools"
)

// InstallComponent コンポーネントをインストールする
func InstallComponent(componentDefineJSON commonpkgtools.ComponentDefineJSON, sourcePath string) bool {
	componentName := componentDefineJSON.ComponentName
	frontendComponentPath := "frontend/src/components/" + componentName
	if componentDefineJSON.Status.HasFrontend {
		wrappedFrontendComponentPath := sourcePath + "/frontend"
		alreadyComponentExists, err := commonpkgtools.DirectoryExists(frontendComponentPath)
		if err != nil {
			log.Println(err)
			return false
		}
		if alreadyComponentExists {
			fmt.Println(componentName + "は既にインストールされています。同名のコンポーネントをアンインストールしてから再度インストールしてください。")
			return false
		}

		err = commonpkgtools.CopyDirectory(wrappedFrontendComponentPath, frontendComponentPath)
		if err != nil {
			log.Println(err)
			return false
		}
	}
	if componentDefineJSON.Status.HasBackend {
		backendComponentPath := "backend/components/" + componentName
		wrappedBackendComponentPath := sourcePath + "/backend"
		alreadyComponentExists, err := commonpkgtools.DirectoryExists(backendComponentPath)
		if err != nil {
			log.Println(err)
			return false
		}
		if alreadyComponentExists {
			fmt.Println(componentName + "は既にインストールされています。同名のコンポーネントをアンインストールしてから再度インストールしてください。")
			if componentDefineJSON.Status.HasFrontend {
				err = commonpkgtools.RemoveDirectory(frontendComponentPath)
				if err != nil {
					fmt.Println("インストールが進行していたフロントエンドコンポーネントの削除に失敗しました。以下コマンドで削除してください。")
					fmt.Println("rm -rf " + frontendComponentPath)
				}
			}
			return false
		}

		err = commonpkgtools.CopyDirectory(backendComponentPath, wrappedBackendComponentPath)
		if err != nil {
			log.Println(err)
			if componentDefineJSON.Status.HasFrontend {
				err = commonpkgtools.RemoveDirectory(frontendComponentPath)
				if err != nil {
					fmt.Println("インストールが進行していたフロントエンドコンポーネントの削除に失敗しました。以下コマンドで削除してください。")
					fmt.Println("rm -rf " + frontendComponentPath)
				}
			}
			return false
		}
	}

	err := commonpkgtools.RemoveDirectory(sourcePath)
	if err != nil {
		fmt.Println("インストール時に使用したディレクトリの削除に失敗しました。必要なら以下コマンドで削除してください。")
		fmt.Println("rm -rf " + sourcePath)
		return false
	}
	return true
}
