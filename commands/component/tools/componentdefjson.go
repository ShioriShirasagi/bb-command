package tools

type ComponentAuthor struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type ComponentStatus struct {
	HasFrontend bool `json:"has_frontend"`
	HasBackend  bool `json:"has_backend"`
}

type CompoentBackendConfig struct {
	GoPkgDepends []string `json:"go_pkg_depends"`
}

type ComponentFrontendConfig struct {
	NpmPackageDepends []string `json:"npm_pkg_depends"`
}

type ComponentDefineJSON struct {
	ComponentName    string                  `json:"component_name"`
	Version          string                  `json:"version"`
	Author           ComponentAuthor         `json:"author"`
	Status           ComponentStatus         `json:"status"`
	ComponentDepends []string                `json:"component_depends"`
	Backend          CompoentBackendConfig   `json:"backend"`
	Frontend         ComponentFrontendConfig `json:"frontend"`
}
