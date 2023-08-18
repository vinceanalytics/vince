package plug

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/tokens"
)

func Auth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearer(r.Header)
		if token == "" || !tokens.Valid(db.Get(r.Context()), token) {
			render.ERROR(w, http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, r)
	})
}
