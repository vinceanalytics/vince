package auth

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/templates"
)

var passwordResetRequestFromTpl = templates.Focus("auth/password_reset_request_form.html")

func PasswordResetRequestForm(w http.ResponseWriter, r *http.Request) {
	render.HTML(r.Context(), w, passwordResetRequestFromTpl, http.StatusOK)
}
