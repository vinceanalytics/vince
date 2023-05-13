package sites

import (
	"net/http"

	"github.com/gernest/vince/render"
)

func CreateSite(w http.ResponseWriter, r *http.Request) {
	render.ERROR(r.Context(), w, http.StatusNotImplemented)
}
