package site

import (
	"net/http"

	"github.com/vinceanalytics/vince/render"
)

func EnableSpikeNotification(w http.ResponseWriter, r *http.Request) {
	render.ERROR(r.Context(), w, http.StatusNotImplemented)
}
