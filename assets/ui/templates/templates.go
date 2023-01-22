package templates

import (
	"embed"
	"html/template"
)

//go:embed layouts auth error
var files embed.FS

var Login = template.Must(
	template.ParseFS(files,
		"layouts/focus.html",
		"auth/login_form.html",
	),
)

var Register = template.Must(
	template.ParseFS(files,
		"layouts/focus.html",
		"auth/register_form.html",
	),
)

var Error = template.Must(template.ParseFS(files,
	"error/error.html",
))

func steps() []string {
	return []string{
		"Register", "Activate account", "Add site info", "Install snippet",
	}
}

func funcs() template.FuncMap {
	return template.FuncMap{
		"steps": steps,
	}
}
