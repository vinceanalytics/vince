package templates

import (
	"embed"
	"html/template"
)

//go:embed layouts
var files embed.FS

var Layouts = template.Must(template.ParseFS(files,
	"layouts/*.html",
))
