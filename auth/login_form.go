package auth

import (
	"net/http"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/render"
)

func LoginForm(w http.ResponseWriter, r *http.Request) {
	render.HTML(r.Context(), w, templates.LoginForm, http.StatusOK)
}
