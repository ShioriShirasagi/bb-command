package install

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ShioriShirasagi/bb-command/commands/component/install/tools"
	commonpkgtools "github.com/ShioriShirasagi/bb-command/commands/component/tools"
	"github.com/ShioriShirasagi/bb-command/commands/version"
)

func Exec(flags ...string) bool {
	if cantContinue(flags) {
		return false
	}

	manifest, err := commonpkgtools.ReadManifest()
	if err != nil {
		log.Println(err)
		return false
	}

	in := newInstaller(manifest)
	if err := in.install(flags[0], nil, true); err != nil {
		log.Println(err)
		in.rollback()
		return false
	}

	if err := in.installPackages(); err != nil {
		log.Println(err)
		fmt.Println("パッケージのインストールに失敗しました。")
		in.rollback()
		return false
	}

	if err := tools.WriteRouting(); err != nil {
		log.Println(err)
		fmt.Println("ルーティングファイルの生成に失敗しました。")
		in.rollback()
		return false
	}

	if err := in.writeManifest(); err != nil {
		log.Println(err)
		fmt.Println(commonpkgtools.ManifestPath + "の書き込みに失敗しました。")
		in.rollback()
		return false
	}

	for _, conflict := range tools.FindRouteConflicts() {
		fmt.Println("警告: " + conflict)
	}

	fmt.Println("コンポーネントのインストールに成功しました。")

	return true
}

type installState int

const (
	stateInstalling installState = iota + 1
	stateInstalled
)

type installedComponent struct {
	key      string
	def      commonpkgtools.ComponentDefineJSON
	explicit bool
}

type installer struct {
	manifest  commonpkgtools.Manifest
	states    map[string]installState
	versions  map[string]string
	installed []installedComponent
	snapshot  *tools.FileSnapshot
	addedNpm  []string
	addedGo   []string
}

func newInstaller(manifest commonpkgtools.Manifest) *installer {
	return &installer{
		manifest: manifest,
		states:   map[string]installState{},
		versions: map[string]string{},
	}
}

// install 依存するコンポーネントをインストールする
func (in *installer) install(spec string, chain []string, isRoot bool) error {
	owner, name, version, err := parseSpec(spec)
	if err != nil {
		return err
	}
	key := owner + "/" + name
	chain = append(chain, key)

	switch in.states[key] {
	case stateInstalling:
		return fmt.Errorf("循環依存を検出しました: %s", strings.Join(chain, " -> "))
	case stateInstalled:
		if prev := in.versions[key]; prev != version {
			return fmt.Errorf("%s のバージョン指定が競合しています (%s と %s)", key, prev, version)
		}
		return nil
	}

	in.states[key] = stateInstalling
	in.versions[key] = version

	if current, ok := in.manifest.Components[key]; ok {
		if isRoot {
			return fmt.Errorf("%s は既にインストールされています", key)
		}
		if version != "latest" && version != current.Version {
			fmt.Printf("警告: %s は %s が要求されていますが、インストール済みの %s を使用します。\n", key, version, current.Version)
		}
		in.states[key] = stateInstalled
		return nil
	}

	workDir, err := os.MkdirTemp("", "bb-install-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(workDir)

	sourcePath, def, err := fetchComponent(owner, name, version, workDir)
	if err != nil {
		return err
	}
	if err := validatePackages(def); err != nil {
		return fmt.Errorf("%s の定義が不正です: %w", key, err)
	}

	for _, dep := range def.ComponentDepends {
		if err := in.install(dep, chain, false); err != nil {
			return err
		}
	}

	if err := tools.InstallComponent(key, def, sourcePath); err != nil {
		return fmt.Errorf("%s のインストールに失敗しました: %w", key, err)
	}
	in.states[key] = stateInstalled
	in.installed = append(in.installed, installedComponent{key: key, def: def, explicit: isRoot})
	return nil
}

// validatePackages 依存パッケージの指定を検証する
func validatePackages(def commonpkgtools.ComponentDefineJSON) error {
	for _, spec := range def.Frontend.NpmPackageDepends {
		if _, _, err := tools.ParseNpmPackage(spec); err != nil {
			return err
		}
	}
	for _, spec := range def.Backend.GoPkgDepends {
		if _, _, err := tools.ParseGoPackage(spec); err != nil {
			return err
		}
	}
	return nil
}

// installPackages 今回インストールしたコンポーネントが依存するパッケージを追加する
func (in *installer) installPackages() error {
	var npmSpecs, goSpecs []string
	for _, c := range in.installed {
		npmSpecs = append(npmSpecs, c.def.Frontend.NpmPackageDepends...)
		goSpecs = append(goSpecs, c.def.Backend.GoPkgDepends...)
	}
	if len(npmSpecs) == 0 && len(goSpecs) == 0 {
		return nil
	}

	snapshot, err := tools.TakeSnapshot(tools.PackageFiles()...)
	if err != nil {
		return err
	}
	in.snapshot = snapshot

	if len(npmSpecs) > 0 {
		in.addedNpm, err = tools.InstallNpmPackages(npmSpecs)
		if err != nil {
			return err
		}
	}
	if len(goSpecs) > 0 {
		in.addedGo, err = tools.InstallGoPackages(goSpecs)
		if err != nil {
			return err
		}
	}
	return nil
}

// rollback 今回インストールしたコンポーネントを削除する
func (in *installer) rollback() {
	if in.snapshot != nil {
		fmt.Println("追加したパッケージを元に戻します。")
		if err := in.snapshot.Restore(); err != nil {
			log.Println(err)
			fmt.Println("パッケージ定義ファイルの復元に失敗しました。package.json、pnpm-lock.yaml、go.mod、go.sumを確認してください。")
		} else if len(in.addedNpm) > 0 {
			if err := tools.SyncNpmPackages(); err != nil {
				log.Println(err)
				fmt.Println("node_modulesの更新に失敗しました。frontendディレクトリで pnpm install を実行してください。")
			}
		}
		in.snapshot = nil
	}

	if len(in.installed) == 0 {
		return
	}

	fmt.Println("インストール済みの依存コンポーネントを削除します。")
	for i := len(in.installed) - 1; i >= 0; i-- {
		tools.UninstallComponent(in.installed[i].key, in.installed[i].def)
	}
	if err := tools.WriteRouting(); err != nil {
		log.Println(err)
		fmt.Println("ルーティングファイルの再生成に失敗しました。routes.tsとroute.goを確認してください。")
	}
	in.installed = nil
}

// writeManifest 今回インストールしたコンポーネントを管理ファイルに記録する
func (in *installer) writeManifest() error {
	for _, c := range in.installed {
		depends := []string{}
		for _, dep := range c.def.ComponentDepends {
			owner, name, _, err := parseSpec(dep)
			if err != nil {
				return err
			}
			depends = append(depends, owner+"/"+name)
		}
		npmPackages := []string{}
		for _, spec := range c.def.Frontend.NpmPackageDepends {
			name, _, err := tools.ParseNpmPackage(spec)
			if err != nil {
				return err
			}
			npmPackages = append(npmPackages, name)
		}
		goPackages := []string{}
		for _, spec := range c.def.Backend.GoPkgDepends {
			path, _, err := tools.ParseGoPackage(spec)
			if err != nil {
				return err
			}
			goPackages = append(goPackages, path)
		}
		in.manifest.Components[c.key] = commonpkgtools.InstalledComponent{
			Version:     c.def.Version,
			HasFrontend: c.def.Status.HasFrontend,
			HasBackend:  c.def.Status.HasBackend,
			Explicit:    c.explicit,
			Depends:     depends,
			NpmPackages: npmPackages,
			GoPackages:  goPackages,
		}
	}
	managed := &in.manifest.ManagedPackages
	managed.Npm = appendMissing(managed.Npm, in.addedNpm)
	managed.Go = appendMissing(managed.Go, in.addedGo)
	in.manifest.BBVersion = version.Version
	return commonpkgtools.WriteManifest(in.manifest)
}

func appendMissing(list, values []string) []string {
	if list == nil {
		list = []string{}
	}
	for _, v := range values {
		if !slices.Contains(list, v) {
			list = append(list, v)
		}
	}
	return list
}

// parseSpec 「owner名/パッケージ名@バージョン」を分解する
func parseSpec(spec string) (owner, name, version string, err error) {
	ownerAndComponentName := strings.Split(spec, "/")
	if len(ownerAndComponentName) != 2 {
		return "", "", "", fmt.Errorf("引数が不正です (%s)。パッケージの指定は「owner名/パッケージ名@バージョン」の形式でなければなりません。", spec)
	}
	owner = ownerAndComponentName[0]

	nameAndVersion := strings.Split(ownerAndComponentName[1], "@")
	name = nameAndVersion[0]
	version = "latest"
	if len(nameAndVersion) >= 2 && nameAndVersion[1] != "" {
		version = nameAndVersion[1]
	}
	if owner == "" || name == "" {
		return "", "", "", fmt.Errorf("引数が不正です (%s)。owner名とパッケージ名は省略できません。", spec)
	}
	return owner, name, version, nil
}

// fetchComponent コンポーネントをworkDirにダウンロード・展開する
func fetchComponent(owner, name, version, workDir string) (string, commonpkgtools.ComponentDefineJSON, error) {
	var def commonpkgtools.ComponentDefineJSON

	downloadURL, fileName, err := tools.FindGitHubComponent(owner, name, version)
	if err != nil {
		return "", def, err
	}

	downloadPath := filepath.Join(workDir, filepath.Base(fileName))
	if err := tools.DownloadFile(downloadURL, downloadPath); err != nil {
		return "", def, err
	}

	sourcePath := filepath.Join(workDir, "source")
	if err := tools.ZipExtract(downloadPath, sourcePath); err != nil {
		return "", def, err
	}

	def, err = commonpkgtools.ReadComponentDefineJSON(sourcePath + "/component_define.json")
	if err != nil {
		return "", def, err
	}
	return sourcePath, def, nil
}

// cantContinue コマンドが続けられるか判定する
func cantContinue(flags []string) bool {
	if len(flags) == 0 {
		fmt.Println("実行に必要な引数が不足しています。")
		return true
	}

	if flags[0] == "help" {
		cmdDetail()
		return true
	}
	return false
}

// cmdDetail コマンド詳細を表示する
func cmdDetail() {
	fmt.Print("Barebornにコンポーネントをインストールします。ShioriShirasagi/SampleComponent@v1.0.0のように、Github名/コンポーネント名@バージョンを引数に取ります。")
}
