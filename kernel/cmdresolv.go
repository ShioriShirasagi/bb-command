package kernel

import (
	"flag"
	"fmt"
	"os"

	"github.com/ShioriShirasagi/bb-command/config"
)

func CommandResolver() {
	flag.Parse()
	args := flag.Args()

	// ヘルプコマンドの存在有無を確認する
	helpCommand, helpCommandExists := config.CommandList["help"]
	if !helpCommandExists {
		// ヘルプコマンドが存在しないことはない
		fmt.Println("エラーが発生しました： 予期せぬ問題が発生しました。コマンドを終了します。")
		os.Exit(1)
	}

	if len(args) < 1 {
		// 引数なしで呼ばれたらヘルプコマンドを起動する
		helpCommand()
		os.Exit(1)
	}

	command := os.Args[1]

	commandFunc, commandFuncExists := config.CommandList[args[0]]
	if !commandFuncExists {
		fmt.Println("エラーが発生しました： コマンド「" + command + "」は存在しません。")
		// 不明なコマンドが呼ばれたらヘルプコマンドを起動する
		helpCommand()
		os.Exit(1)
	}

	// 先頭のコマンド名（pkg:wrap等）を除いてコマンド実行
	success := commandFunc(args[1:]...)
	if success {
		os.Exit(0)
	}
	os.Exit(1)
}
