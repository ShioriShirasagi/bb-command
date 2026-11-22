package install

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/ShioriShirasagi/bb-command/commands/component/install/tools"
	commonpkgtools "github.com/ShioriShirasagi/bb-command/commands/component/tools"
)

func Exec(flags ...string) bool {
	if cantContinue(flags) {
		return false
	}

	componentSpecify := flags[0]
	ownerAndComponentName := strings.Split(componentSpecify, "/")

	if len(ownerAndComponentName) != 2 {
		fmt.Println("引数が不正です。パッケージの指定は「owner名/パッケージ名@バージョン」の形式でなければなりません。")
		return false
	}
	componentOwner := ownerAndComponentName[0]
	componentNameWithVersion := ownerAndComponentName[1]
	componentNameAndVersion := strings.Split(componentNameWithVersion, "@")
	componentName := ""
	componentVersion := "latest"
	if len(componentNameAndVersion) == 1 {
		componentName = componentNameWithVersion
	} else {
		componentName = componentNameAndVersion[0]
		componentVersion = componentNameAndVersion[1]
	}

	downloadURL, fileName, err := tools.FindGitHubComponent(componentOwner, componentName, componentVersion)
	if err != nil {
		log.Println(err)
		return false
	}

	downloadPath := filepath.Base(fileName)
	err = tools.DownloadFile(downloadURL, downloadPath)
	if err != nil {
		log.Println(err)
		return false
	}

	err = tools.ZipExtract(downloadPath, componentName)
	if err != nil {
		log.Println(err)
		return false
	}

	componentDefJSONPath := componentName + "/component_define.json"
	componentDefineJSON, err := commonpkgtools.ReadComponentDefineJSON(componentDefJSONPath)
	if err != nil {
		log.Println(err)
		return false
	}

	if !tools.InstallComponent(componentDefineJSON, componentName) {
		return false
	}

	fmt.Println("コンポーネントのインストールに成功しました。")

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
	fmt.Print("Barebornにコンポーネントをインストールします。ShioriShirasagi/SampleComponent@v1.0.0のように、Github名/コンポーネント名@バージョンを引数に取ります。")
}
