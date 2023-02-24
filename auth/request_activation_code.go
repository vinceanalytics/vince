package auth

import (
	"net/http"

	"github.com/gernest/vince/render"
)

func RequestActivationCode(w http.ResponseWriter, r *http.Request) {
	render.ERROR(r.Context(), w, http.StatusNotImplemented)
}
