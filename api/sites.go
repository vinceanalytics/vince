package api

import (
	"net/http"

	"github.com/gernest/vince/render"
)

func Sites(w http.ResponseWriter, r *http.Request) {
	render.ERROR(r.Context(), w, http.StatusNotImplemented)
}
