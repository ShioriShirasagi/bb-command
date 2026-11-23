package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"
)

const (
	frontendDir = "frontend"
	backendDir  = "backend"
)

// コマンドの引数として安全な文字だけを許可する
var (
	npmNamePattern    = regexp.MustCompile(`^(@[a-z0-9][a-z0-9._~-]*/)?[a-z0-9][a-z0-9._~-]*$`)
	npmVersionPattern = regexp.MustCompile(`^[0-9A-Za-z.^~*<>=|+][0-9A-Za-z.^~*<>=|+-]*$`)
	goPathPattern     = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*(\.[a-z0-9-]+)+(/[A-Za-z0-9_~+-][A-Za-z0-9._~+-]*)*$`)
	goVersionPattern  = regexp.MustCompile(`^(latest|v[0-9]+\.[0-9]+\.[0-9]+([-+][0-9A-Za-z.-]+)?)$`)
)

// ParseNpmPackage 「パッケージ名@バージョン」を検証して分解する
func ParseNpmPackage(spec string) (name, version string, err error) {
	name = spec
	if i := strings.LastIndex(spec, "@"); i > 0 {
		name, version = spec[:i], spec[i+1:]
		if !npmVersionPattern.MatchString(version) {
			return "", "", fmt.Errorf("npmパッケージのバージョン指定が不正です: %q", spec)
		}
	}
	if len(name) > 214 || !npmNamePattern.MatchString(name) {
		return "", "", fmt.Errorf("npmパッケージ名が不正です: %q", spec)
	}
	return name, version, nil
}

// ParseGoPackage 「モジュールパス@バージョン」を検証して分解する
func ParseGoPackage(spec string) (path, version string, err error) {
	path, version, _ = strings.Cut(spec, "@")
	if version == "" {
		version = "latest"
	}
	if !goPathPattern.MatchString(path) || strings.Contains(path, "./") || strings.HasSuffix(path, ".") {
		return "", "", fmt.Errorf("Goモジュールのパスが不正です: %q", spec)
	}
	if !goVersionPattern.MatchString(version) {
		return "", "", fmt.Errorf("Goモジュールのバージョン指定が不正です: %q", spec)
	}
	return path, version, nil
}

// InstallNpmPackages 未インストールのnpmパッケージを追加し、追加したパッケージ名を返す
func InstallNpmPackages(specs []string) ([]string, error) {
	existing, err := npmDependencies()
	if err != nil {
		return nil, err
	}

	var names, args []string
	for _, spec := range specs {
		name, version, err := ParseNpmPackage(spec)
		if err != nil {
			return nil, err
		}
		if _, ok := existing[name]; ok || slices.Contains(names, name) {
			fmt.Printf("npmパッケージ %s は既に追加されているためスキップします。\n", name)
			continue
		}
		names = append(names, name)
		if version != "" {
			args = append(args, name+"@"+version)
		} else {
			args = append(args, name)
		}
	}
	if len(args) == 0 {
		return nil, nil
	}

	return names, runCommand(frontendDir, "pnpm", append([]string{"add"}, args...)...)
}

// InstallGoPackages 未インストールのGoモジュールを追加し、追加したモジュールパスを返す
func InstallGoPackages(specs []string) ([]string, error) {
	existing, err := goRequires()
	if err != nil {
		return nil, err
	}

	var paths, args []string
	for _, spec := range specs {
		path, version, err := ParseGoPackage(spec)
		if err != nil {
			return nil, err
		}
		if existing[path] || slices.Contains(paths, path) {
			fmt.Printf("Goモジュール %s は既に追加されているためスキップします。\n", path)
			continue
		}
		paths = append(paths, path)
		args = append(args, path+"@"+version)
	}
	if len(args) == 0 {
		return nil, nil
	}

	return paths, runCommand(backendDir, "go", append([]string{"get"}, args...)...)
}

// SyncNpmPackages package.jsonに合わせてnode_modulesを更新する
func SyncNpmPackages() error {
	return runCommand(frontendDir, "pnpm", "install")
}

func npmDependencies() (map[string]string, error) {
	raw, err := os.ReadFile(frontendDir + "/package.json")
	if err != nil {
		return nil, err
	}
	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(raw, &pkg); err != nil {
		return nil, err
	}

	deps := map[string]string{}
	for name, version := range pkg.DevDependencies {
		deps[name] = version
	}
	for name, version := range pkg.Dependencies {
		deps[name] = version
	}
	return deps, nil
}

func goRequires() (map[string]bool, error) {
	cmd := exec.Command("go", "mod", "edit", "-json")
	cmd.Dir = backendDir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go.modの読み込みに失敗しました: %w", err)
	}
	var mod struct {
		Require []struct {
			Path string `json:"Path"`
		} `json:"Require"`
	}
	if err := json.Unmarshal(out, &mod); err != nil {
		return nil, err
	}

	requires := map[string]bool{}
	for _, r := range mod.Require {
		requires[r.Path] = true
	}
	return requires, nil
}

// runCommand シェルを介さずにコマンドを実行する
func runCommand(dir, name string, args ...string) error {
	if _, err := exec.LookPath(name); err != nil {
		return fmt.Errorf("%s が見つかりません: %w", name, err)
	}
	fmt.Printf("実行中: %s %s\n", name, strings.Join(args, " "))

	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s の実行に失敗しました: %w", name, err)
	}
	return nil
}

type FileSnapshot struct {
	contents map[string][]byte
}

// TakeSnapshot ファイルの内容を保存する
func TakeSnapshot(paths ...string) (*FileSnapshot, error) {
	snapshot := &FileSnapshot{contents: map[string][]byte{}}
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if errors.Is(err, fs.ErrNotExist) {
			snapshot.contents[path] = nil
			continue
		}
		if err != nil {
			return nil, err
		}
		snapshot.contents[path] = content
	}
	return snapshot, nil
}

// Restore 保存した内容にファイルを戻す
func (s *FileSnapshot) Restore() error {
	var errs []error
	for path, content := range s.contents {
		if content == nil {
			if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
				errs = append(errs, err)
			}
			continue
		}
		if err := os.WriteFile(path, content, 0o644); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// PackageFiles パッケージ追加で変更されるファイル
func PackageFiles() []string {
	return []string{
		frontendDir + "/package.json",
		frontendDir + "/pnpm-lock.yaml",
		backendDir + "/go.mod",
		backendDir + "/go.sum",
	}
}
