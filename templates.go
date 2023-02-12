package vince

import (
	"context"
	"html/template"
	"net/http"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/log"
)

func ServeHTML(ctx context.Context, w http.ResponseWriter, tpl *template.Template, code int, data any) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(code)
	err := tpl.Execute(w, data)
	if err != nil {
		log.Get(ctx).Err(err).Str("template", tpl.Name()).Msg("Failed to render")
	}
}

func ServeError(ctx context.Context, w http.ResponseWriter, code int) {
	ServeHTML(ctx, w, templates.Error, code, map[string]any{
		"Status": code,
		"Text":   http.StatusText(code),
	})
}
