package wrap

import (
	"fmt"
	"log"

	commonpkgtools "github.com/ShioriShirasagi/bb-command/commands/component/tools"
	"github.com/ShioriShirasagi/bb-command/commands/component/wrap/tools"
)

func Exec(flags ...string) bool {
	if cantContinue(flags) {
		return false
	}

	componentName := flags[0]
	componentDefJSONPath := "component/define/" + componentName + ".bbcomp.json"
	componentWrapDir := "homemade/" + componentName

	// コンポーネント定義JSONの存在確認
	fileExists := commonpkgtools.FileExists(componentDefJSONPath)
	if !fileExists {
		fmt.Println(componentName + "のコンポーネント定義JSONが見つかりません。")
		return false
	}

	componentDefineJSON, err := commonpkgtools.ReadComponentDefineJSON(componentDefJSONPath)
	if err != nil {
		log.Println(err)
		return false
	}

	alreadyPackageExists, err := tools.DirectoryExists(componentWrapDir)
	if err != nil {
		log.Println(err)
		return false
	}

	if alreadyPackageExists {
		fmt.Println(componentName + "のディレクトリは既に存在します。")
		return false
	}

	tools.FolderGenerator(componentWrapDir)

	if componentDefineJSON.Status.HasFrontend {
		frontendComponentPath := "frontend/src/components/" + componentName
		wrappedFrontendComponentPath := componentWrapDir + "/frontend"
		err := tools.CopyDirectory(frontendComponentPath, wrappedFrontendComponentPath)
		if err != nil {
			log.Println(err)
			return false
		}
	}
	if componentDefineJSON.Status.HasBackend {
		backendComponentPath := "backend/components/" + componentName
		wrappedBackendComponentPath := componentWrapDir + "/backend"
		err := tools.CopyDirectory(backendComponentPath, wrappedBackendComponentPath)
		if err != nil {
			log.Println(err)
			return false
		}
	}

	wrappedComponentJSONPath := componentWrapDir + "/component_define.json"
	err = tools.CopyFile(componentDefJSONPath, wrappedComponentJSONPath)
	if err != nil {
		log.Println(err)
		return false
	}

	tools.TarGzArchive(componentWrapDir, componentWrapDir+".tar.gz")
	err = commonpkgtools.RemoveDirectory(componentWrapDir)
	if err != nil {
		fmt.Println("アーカイブ作成中に発生したディレクトリの削除に失敗しました。必要であれば以下コマンドで削除してください。")
		fmt.Println("rm -rf " + componentWrapDir)
	}

	fmt.Println("「" + componentWrapDir + ".tar.gz」にアーカイブの作成が完了しました。")

	return true
}

// cantContinue コマンドが続けられるか判定する
func cantContinue(flags []string) bool {
	if len(flags) == 0 {
		fmt.Println("実行に必要な引数が不足しています。")
		return true
	}

	if flags[0] == "help" {
		cmdDetail()
		return true
	}
	return false
}

// cmdDetail コマンド詳細を表示する
func cmdDetail() {
	fmt.Print("開発したパッケージをまとめて頒布可能な形式にします。コンポーネント名を引数に取ります。")
}
