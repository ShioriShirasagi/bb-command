package tools

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/ShioriShirasagi/bb-command/helper"
)

var echoMethods = map[string]bool{
	"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true,
	"HEAD": true, "OPTIONS": true, "CONNECT": true, "TRACE": true, "Any": true,
}

var frontendPathPattern = regexp.MustCompile(`path\s*:\s*['"` + "`" + `]([^'"` + "`" + `]+)['"` + "`" + `]`)

// FindRouteConflicts コンポーネント間で重複しているURLを探す
// 文字列で直接書かれたURLしか検出できない
func FindRouteConflicts() []string {
	var conflicts []string
	conflicts = append(conflicts, findBackendConflicts()...)
	conflicts = append(conflicts, findFrontendConflicts()...)
	return conflicts
}

func findBackendConflicts() []string {
	dirs, err := helper.ComponentDirs(backendComponentsRoot)
	if err != nil {
		return nil
	}

	owners := map[string][]string{}
	for _, dir := range dirs {
		f, err := parser.ParseFile(token.NewFileSet(), filepath.Join(backendComponentsRoot, dir, "route.go"), nil, 0)
		if err != nil {
			continue
		}
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) == 0 {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || !echoMethods[sel.Sel.Name] {
				return true
			}
			lit, ok := call.Args[0].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			path, err := strconv.Unquote(lit.Value)
			if err != nil {
				return true
			}
			owners[path] = appendUnique(owners[path], dir)
			return true
		})
	}
	return conflictMessages("バックエンド", owners)
}

func findFrontendConflicts() []string {
	dirs, err := helper.ComponentDirs(frontendComponentsRoot)
	if err != nil {
		return nil
	}

	owners := map[string][]string{}
	for _, dir := range dirs {
		src, err := os.ReadFile(filepath.Join(frontendComponentsRoot, dir, "router/route.ts"))
		if err != nil {
			continue
		}
		for _, m := range frontendPathPattern.FindAllSubmatch(src, -1) {
			path := string(m[1])
			owners[path] = appendUnique(owners[path], dir)
		}
	}
	return conflictMessages("フロントエンド", owners)
}

func appendUnique(list []string, v string) []string {
	for _, s := range list {
		if s == v {
			return list
		}
	}
	return append(list, v)
}

func conflictMessages(side string, owners map[string][]string) []string {
	var messages []string
	for path, dirs := range owners {
		if len(dirs) >= 2 {
			messages = append(messages, fmt.Sprintf("%sのURL「%s」が重複しています: %s", side, path, strings.Join(dirs, ", ")))
		}
	}
	sort.Strings(messages)
	return messages
}
