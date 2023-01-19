package vince

import (
	"net/http"

	"github.com/gernest/vince/assets/ui/templates"
)

func (v *Vince) login() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ServeHTML(w, templates.Login, http.StatusOK, map[string]any{})
	})
}
