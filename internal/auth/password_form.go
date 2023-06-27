package auth

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/templates"
)

var passwordFormTpl = templates.Focus("auth/password_form.html")

func PasswordForm(w http.ResponseWriter, r *http.Request) {
	render.HTML(r.Context(), w, passwordFormTpl, http.StatusOK)
}
