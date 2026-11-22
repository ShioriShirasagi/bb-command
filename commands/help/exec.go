package help

import (
	"fmt"
)

var Commands map[string]func(flags ...string)

func Exec(flags ...string) {
	helpTxt := `BBコマンドヘルプ
基本： ./bb <Command> <Flags>
コマンド一覧：
`
	fmt.Print(helpTxt)
	for commandName, command := range Commands {
		if commandName == "help" {
			continue
		}
		fmt.Print(commandName + " -> ")
		command("help")
	}
}
