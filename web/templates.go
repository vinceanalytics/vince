package web

import (
	"embed"
	"html/template"
)

//go:embed templates
var templateData embed.FS

var (
	layouts = template.Must(template.ParseFS(
		templateData, "templates/layout/*",
	))
	home = template.Must(layouts.Lookup("focus").ParseFS(templateData, "templates/page/index.html"))
)
