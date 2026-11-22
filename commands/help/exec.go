package help

import (
	"fmt"
)

var Commands map[string]func(flags ...string) bool

func Exec(flags ...string) bool {
	helpTxt := `BBコマンドヘルプ
基本： ./bb <Command> <Flags>

コマンド一覧：
`
	fmt.Print(helpTxt)
	for commandName, command := range Commands {
		if commandName == "help" {
			fmt.Println("・help -> BBコマンドのヘルプを表示します。")
			continue
		}
		fmt.Print("・" + commandName + " -> ")
		command("help")
		fmt.Println("")
	}

	return true
}
