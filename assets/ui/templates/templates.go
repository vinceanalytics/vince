package templates

import (
	"embed"
	"html/template"

	"github.com/belak/octicon"
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
	template.Must(template.ParseFS(files,
		"layouts/focus.html",
	)).Funcs(funcs).ParseFS(files,
		"auth/register_form.html",
	),
)

var Error = template.Must(template.ParseFS(files,
	"error/error.html",
))

var funcs = template.FuncMap{
	"oAlertFill": wrap(octicon.AlertFill),
}

func wrap(f func(int, ...string) (string, bool)) func(int) template.HTML {
	return func(i int) template.HTML {
		v, _ := f(i)
		return template.HTML(v)
	}
}
