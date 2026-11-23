// templates/templates.go
package templates

import _ "embed"

//go:embed routefile.go.tmpl
var Routes string

//go:embed routefile.ts.tmpl
var FrontendRoutes string
