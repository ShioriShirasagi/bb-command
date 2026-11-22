package config

import (
	"github.com/ShioriShirasagi/bb-command/commands/component/wrap"
	"github.com/ShioriShirasagi/bb-command/commands/help"
	"github.com/ShioriShirasagi/bb-command/commands/version"
)

// CommandList 実行可能なコマンド一覧
var CommandList = map[string]func(flags ...string) bool{
	"version":        version.Exec,
	"help":           help.Exec,
	"component:wrap": wrap.Exec,
}

func init() {
	help.Commands = CommandList
}
