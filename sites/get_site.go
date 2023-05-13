package sites

import (
	"net/http"

	"github.com/gernest/vince/render"
)

func GetSite(w http.ResponseWriter, r *http.Request) {
	render.ERROR(r.Context(), w, http.StatusNotImplemented)
}
