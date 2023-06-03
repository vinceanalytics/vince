package auth

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/templates"
)

func PasswordResetRequestForm(w http.ResponseWriter, r *http.Request) {
	render.HTML(r.Context(), w, templates.PasswordResetRequestForm, http.StatusOK)
}
