package helper

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type generateTarget struct {
	Alias      string
	ImportPath string
}

// ComponentDirs root配下の「owner/name」形式のディレクトリを列挙する
func ComponentDirs(root string) ([]string, error) {
	owners, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	var dirs []string
	for _, owner := range owners {
		if !owner.IsDir() {
			continue
		}
		names, err := os.ReadDir(filepath.Join(root, owner.Name()))
		if err != nil {
			return nil, err
		}
		for _, name := range names {
			if name.IsDir() {
				dirs = append(dirs, owner.Name()+"/"+name.Name())
			}
		}
	}
	return dirs, nil
}

// importAlias 「owner/name」からimport用の別名を作る
func importAlias(dir string) (string, error) {
	owner, name, _ := strings.Cut(dir, "/")
	replacer := strings.NewReplacer("-", "_", ".", "_")
	alias := replacer.Replace(owner) + "__" + replacer.Replace(name)
	if alias[0] >= '0' && alias[0] <= '9' {
		alias = "_" + alias
	}
	if !token.IsIdentifier(alias) {
		return "", fmt.Errorf("%s は識別子として使えない名前です", dir)
	}
	return alias, nil
}

// GenerateGoFile rootを走査してfuncNameを持つコンポーネントを読み込むGoファイルを生成する
func GenerateGoFile(tmplSrc, root, importRoot, fileName, funcName, outPath string) error {
	dirs, err := ComponentDirs(root)
	if err != nil {
		return err
	}

	var found []string
	for _, dir := range dirs {
		f, err := parser.ParseFile(token.NewFileSet(), filepath.Join(root, dir, fileName), nil, 0)
		if err != nil {
			continue // 対象ファイルがなければ読み飛ばす
		}
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if ok && fn.Name.Name == funcName && fn.Recv == nil {
				found = append(found, dir)
				break
			}
		}
	}

	return generate(tmplSrc, found, func(dir string) string { return importRoot + "/" + dir }, outPath, format.Source)
}

// GenerateTSFile rootを走査してfileNameを持つコンポーネントを読み込むTSファイルを生成する
func GenerateTSFile(tmplSrc, root, fileName, outPath string) error {
	dirs, err := ComponentDirs(root)
	if err != nil {
		return err
	}

	var found []string
	for _, dir := range dirs {
		if _, err := os.Stat(filepath.Join(root, dir, fileName)); err == nil {
			found = append(found, dir)
		}
	}

	importFile := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	return generate(tmplSrc, found, func(dir string) string { return "./" + dir + "/" + importFile }, outPath, nil)
}

func generate(tmplSrc string, dirs []string, importPath func(string) string, outPath string, formatter func([]byte) ([]byte, error)) error {
	var targets []generateTarget
	aliases := map[string]string{}
	for _, dir := range dirs {
		alias, err := importAlias(dir)
		if err != nil {
			return err
		}
		if prev, ok := aliases[alias]; ok {
			return fmt.Errorf("%s と %s の別名が重複しています", prev, dir)
		}
		aliases[alias] = dir
		targets = append(targets, generateTarget{Alias: alias, ImportPath: importPath(dir)})
	}

	t, err := template.New("gen").Parse(tmplSrc)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, targets); err != nil {
		return err
	}

	out := buf.Bytes()
	if formatter != nil {
		if out, err = formatter(out); err != nil {
			return err
		}
	}
	return os.WriteFile(outPath, out, 0o644)
}
