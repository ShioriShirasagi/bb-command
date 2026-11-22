package install

import (
	"fmt"
	"log"
	"strings"

	"github.com/ShioriShirasagi/bb-command/commands/component/install/tools"
	commonpkgtools "github.com/ShioriShirasagi/bb-command/commands/component/tools"
)

func Exec(flags ...string) bool {
	if cantContinue(flags) {
		return false
	}

	componentName := flags[0]
	ownerAndRepository := strings.Split(componentName, "/")

	if len(ownerAndRepository) != 2 {
		fmt.Println("引数が不正です。パッケージの指定は「owner名/パッケージ名」でなければなりません。")
	}
	componentOwner := ownerAndRepository[0]
	componentRepository := ownerAndRepository[1]

	downloadPath, err := tools.GitHubPackageDownloader(componentOwner, componentRepository)
	if err != nil {
		log.Println(err)
		return false
	}

	err = tools.ZipExtract(downloadPath, componentRepository)
	if err != nil {
		log.Println(err)
		return false
	}

	componentDefJSONPath := componentRepository + "/component_define.json"
	componentDefineJSON, err := commonpkgtools.ReadComponentDefineJSON(componentDefJSONPath)

	if !tools.InstallComponent(componentDefineJSON, componentRepository) {
		return false
	}

	fmt.Println("コンポーネントのインストールに成功しました！")

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
	fmt.Print("Barebornにコンポーネントをインストールします。ShioriShirasagi/SampleComponentのように、Github名/コンポーネント名を引数に取ります。")
}
