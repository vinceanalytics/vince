package router

import (
	"net/http"

	"github.com/gernest/vince/render"
)

func SubmitErrorReport(w http.ResponseWriter, r *http.Request) {
	render.ERROR(r.Context(), w, http.StatusNotImplemented)
}
