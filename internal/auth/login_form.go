package auth

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/render"
)

func LoginForm(w http.ResponseWriter, r *http.Request) {
	render.HTML(r.Context(), w, loginForm, http.StatusOK)
}
