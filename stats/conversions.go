package stats

import (
	"net/http"

	"github.com/gernest/vince/render"
)

func Conversions(w http.ResponseWriter, r *http.Request) {
	render.ERROR(r.Context(), w, http.StatusNotImplemented)
}
