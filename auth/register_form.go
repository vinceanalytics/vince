package auth

import (
	"net/http"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/render"
)

func RegisterForm(w http.ResponseWriter, r *http.Request) {
	render.HTML(r.Context(), w, templates.RegisterForm, http.StatusOK)
}
