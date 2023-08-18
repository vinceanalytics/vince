package plug

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/render"
)

func AcceptJSON(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("content-type") != "application/json" {
			render.ERROR(w, http.StatusBadRequest)
			return
		}
		h.ServeHTTP(w, r)
	})
}
