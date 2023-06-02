package share

import (
	"net/http"

	"github.com/vinceanalytics/vince/render"
)

func SharedLink(w http.ResponseWriter, r *http.Request) {
	render.ERROR(r.Context(), w, http.StatusNotImplemented)
}
