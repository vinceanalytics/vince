package templates

import (
	"embed"
	"html/template"
)

//go:embed layouts auth error
var files embed.FS

var Layouts = template.Must(template.ParseFS(files,
	"layouts/*.html",
))

var Login = template.Must(
	Layouts.Lookup("focus.html").ParseFS(files, "auth/login_form.html"),
)

var Register = template.Must(
	Layouts.Lookup("focus.html").ParseFS(files, "auth/register_form.html"),
)

var Error = template.Must(template.ParseFS(files,
	"error/error.html",
))
