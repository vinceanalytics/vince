package tracker

import (
	"embed"
	"net/http"
	"strings"

	"github.com/gernest/vince/plug"
)

//go:embed js/*.js
var files embed.FS

func Plug() plug.Plug {
	fs := http.FileServer(http.FS(files))
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/js/vince") {
				w.Header().Set("content-type", "application/javascript")
				w.Header().Set("x-content-type-options", "nosniff")
				w.Header().Set("cross-origin-resource-policy", "cross-origin")
				w.Header().Set("access-control-allow-origin", "*")
				w.Header().Set("cache-control", "public, max-age=86400, must-revalidate")
				fs.ServeHTTP(w, r)
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}
