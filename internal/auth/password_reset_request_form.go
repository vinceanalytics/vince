package auth

import (
	"net/http"

	"github.com/vinceanalytics/vince/render"
)

func PasswordResetRequestForm(w http.ResponseWriter, r *http.Request) {
	render.ERROR(r.Context(), w, http.StatusNotImplemented)
}
