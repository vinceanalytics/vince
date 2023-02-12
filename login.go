package vince

import (
	"net/http"

	"github.com/gernest/vince/assets/ui/templates"
)

func (v *Vince) loginForm() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ServeHTML(r.Context(), w, templates.Login, http.StatusOK, templates.New(r.Context()))
	})
}
