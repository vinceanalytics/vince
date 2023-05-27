package auth

import (
	"net/http"

	"github.com/gernest/vince/render"
	"github.com/gernest/vince/templates"
)

func RegisterForm(w http.ResponseWriter, r *http.Request) {
	render.HTML(r.Context(), w, templates.RegisterForm, http.StatusOK)
}
