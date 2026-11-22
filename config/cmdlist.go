package config

import (
	"github.com/ShioriShirasagi/bb-command/commands/pkg/help"
	"github.com/ShioriShirasagi/bb-command/commands/pkg/wrap"
	"github.com/ShioriShirasagi/bb-command/commands/version"
)

// CommandList 実行可能なコマンド一覧
var CommandList = map[string]func(flags ...string){
	"version":  version.Exec,
	"help":     help.Exec,
	"pkg:wrap": wrap.Exec,
}

func init() {
	help.Commands = CommandList
}
