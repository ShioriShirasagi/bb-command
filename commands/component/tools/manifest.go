package tools

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
)

const ManifestPath = "bbstate.json"

type InstalledComponent struct {
	Version     string   `json:"version"`
	HasFrontend bool     `json:"has_frontend"`
	HasBackend  bool     `json:"has_backend"`
	DevOnly     bool     `json:"dev_only"`
	Explicit    bool     `json:"explicit"`
	Depends     []string `json:"depends"`
	NpmPackages []string `json:"npm_packages"`
	GoPackages  []string `json:"go_packages"`
}

type ManagedPackages struct {
	Npm []string `json:"npm"`
	Go  []string `json:"go"`
}

type Manifest struct {
	BBVersion       string                        `json:"bbversion"`
	Components      map[string]InstalledComponent `json:"components"`
	ManagedPackages ManagedPackages               `json:"managed_packages"`
}

// ReadManifest インストール管理ファイルを読み込む
func ReadManifest() (Manifest, error) {
	manifest := Manifest{Components: map[string]InstalledComponent{}}

	rawJSON, err := os.ReadFile(ManifestPath)
	if errors.Is(err, fs.ErrNotExist) {
		return manifest, nil
	}
	if err != nil {
		return manifest, err
	}

	if err := json.Unmarshal(rawJSON, &manifest); err != nil {
		return manifest, err
	}
	if manifest.Components == nil {
		manifest.Components = map[string]InstalledComponent{}
	}
	return manifest, nil
}

// WriteManifest インストール管理ファイルを書き込む
func WriteManifest(manifest Manifest) error {
	rawJSON, err := json.MarshalIndent(manifest, "", "    ")
	if err != nil {
		return err
	}

	tmp := ManifestPath + ".tmp"
	if err := os.WriteFile(tmp, append(rawJSON, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, ManifestPath)
}
