package sites

import (
	"net/http"

	"github.com/gernest/vince/render"
)

func Get(w http.ResponseWriter, r *http.Request) {
	render.ERROR(r.Context(), w, http.StatusNotImplemented)
}
