package api

import (
	"net/http"

	"github.com/gernest/vince/render"
)

func DomainStatus(w http.ResponseWriter, r *http.Request) {
	render.ERROR(r.Context(), w, http.StatusNotImplemented)
}
