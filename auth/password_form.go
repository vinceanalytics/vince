package auth

import (
	"net/http"

	"github.com/vinceanalytics/vince/render"
	"github.com/vinceanalytics/vince/templates"
)

func PasswordForm(w http.ResponseWriter, r *http.Request) {
	render.HTML(r.Context(), w, templates.PasswordForm, http.StatusOK)
}
