package auth

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/render"
)

func RegisterForm(w http.ResponseWriter, r *http.Request) {
	render.HTML(r.Context(), w, registerForm, http.StatusOK)
}
