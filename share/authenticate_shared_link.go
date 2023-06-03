package share

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/render"
)

func AuthenticateSharedLink(w http.ResponseWriter, r *http.Request) {
	render.ERROR(r.Context(), w, http.StatusNotImplemented)
}
