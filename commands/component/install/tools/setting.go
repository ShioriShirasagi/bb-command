package tools

import (
	"github.com/ShioriShirasagi/bb-command/helper"
	"github.com/ShioriShirasagi/bb-command/templates"
)

const (
	frontendComponentsRoot = "frontend/src/components"
	backendComponentsRoot  = "backend/components"
)

// WriteFrontendRouting フロントエンドのルーティングを生成する
func WriteFrontendRouting() error {
	return helper.GenerateTSFile(
		templates.FrontendRoutes,
		frontendComponentsRoot,
		"router/route.ts",
		frontendComponentsRoot+"/routes.ts",
	)
}

// WriteBackendRouting バックエンドのルーティングを生成する
func WriteBackendRouting() error {
	return helper.GenerateGoFile(
		templates.Routes,
		backendComponentsRoot,
		"github.com/ShioriShirasagi/bareborn-backend/components",
		"route.go",
		"Routes",
		"backend/kernel/internal/provider/route.go",
	)
}

// WriteRouting フロントエンドとバックエンドのルーティングを生成する
func WriteRouting() error {
	if err := WriteFrontendRouting(); err != nil {
		return err
	}
	return WriteBackendRouting()
}
