package vince

import (
	"net/http"

	"github.com/gernest/vince/assets/ui/templates"
)

func (v *Vince) registerForm() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ServeHTML(w, templates.Register, http.StatusOK, map[string]any{})
	})
}
