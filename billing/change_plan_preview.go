package billing

import (
	"net/http"

	"github.com/gernest/vince/render"
)

func ChangePlanPreview(w http.ResponseWriter, r *http.Request) {
	render.ERROR(r.Context(), w, http.StatusNotImplemented)
}
