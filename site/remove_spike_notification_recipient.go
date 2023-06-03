package site

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/render"
)

func RemoveSpikeNotificationRecipient(w http.ResponseWriter, r *http.Request) {
	render.ERROR(r.Context(), w, http.StatusNotImplemented)
}
