package templates

import (
	"embed"
	"html/template"
)

//go:embed layouts auth
var files embed.FS

var Layouts = template.Must(template.ParseFS(files,
	"layouts/*.html",
))

var Login = template.Must(
	Layouts.Lookup("focus.html").ParseFS(files, "auth/login_form.html"),
)
