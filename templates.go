package vince

import (
	"net/http"

	"github.com/gernest/vince/assets/ui/templates"
)

func ServeHTML(w http.ResponseWriter, code int, name string, data any) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(code)
	err := templates.Render(w, name, data)
	if err != nil {
		xlg.Err(err).Str("template", name).Msg("Failed to render")
	}
}
