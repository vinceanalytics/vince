package billing

import (
	"net/http"

	"github.com/gernest/vince/render"
)

func PingSubscription(w http.ResponseWriter, r *http.Request) {
	render.ERROR(r.Context(), w, http.StatusNotImplemented)
}
