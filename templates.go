package vince

import (
	"html/template"
	"net/http"
)

func ServeHTML(w http.ResponseWriter, tpl *template.Template, code int, data any) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(code)
	err := tpl.Execute(w, data)
	if err != nil {
		xlg.Err(err).Str("template", tpl.Name()).Msg("Failed to render")
	}
}
