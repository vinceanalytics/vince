package tracker

import (
	_ "embed"
	"io"
	"text/template"
)

//go:embed track.tmpl
var data string

var tpl = template.Must(
	template.New("").Delims("<<", ">>").Funcs(template.FuncMap{
		"hasCustomEvents": func(a map[string]bool) bool {
			for _, v := range []string{"outbound_links", "file_downloads", "tagged_events"} {
				if a[v] {
					return true
				}
			}
			return false
		},
	}).Parse(string(data)),
)

// Render uses feature to generate  tracker js script.
func Render(w io.Writer, features map[string]bool) error {
	return tpl.Execute(w, features)
}
