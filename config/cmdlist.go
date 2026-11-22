package config

import (
	"github.com/ShioriShirasagi/bb-command/commands/help"
	"github.com/ShioriShirasagi/bb-command/commands/version"
)

// CommandList 実行可能なコマンド一覧
var CommandList = map[string]func(flags ...string){
	"version": version.Exec,
	"help":    help.Exec,
}

func init() {
	help.Commands = CommandList
}
