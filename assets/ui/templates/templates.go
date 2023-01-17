package templates

import (
	"embed"
	"html/template"
	"io"
)

//go:embed layouts
var files embed.FS

var tpl *template.Template

func init() {
	var err error
	tpl, err = template.ParseFS(files,
		"layouts/*.html",
	)
	if err != nil {
		panic(err.Error())
	}
}

func Render(out io.Writer, name string, data any) error {
	return tpl.ExecuteTemplate(out, name, data)
}
