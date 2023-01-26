package templates

import (
	"embed"
	"fmt"
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
	"fieldError": formError,
}

func formError(field string, ctx any) template.HTML {
	a, ok := ctx.(map[string]string)
	if !ok {
		return template.HTML("")
	}
	v, ok := a[field]
	if !ok {
		return template.HTML("")
	}
	errorTpl := `<div class="FormControl-inlineValidation">
	        %s
            <span>%s</span>
        </div>`
	fill, _ := octicon.AlertFill(12)
	return template.HTML(fmt.Sprintf(errorTpl, fill, v))
}
