package tracker

import (
	"embed"
	"net/http"
	"strings"
)

//go:embed js/*.js
var files embed.FS

func Match(s string) bool {
	return strings.HasPrefix(s, "/js/vince")
}

func Serve() http.HandlerFunc {
	h := http.FileServer(http.FS(files))
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/javascript")
		w.Header().Set("x-content-type-options", "nosniff")
		w.Header().Set("cross-origin-resource-policy", "cross-origin")
		w.Header().Set("access-control-allow-origin", "*")
		w.Header().Set("cache-control", "public, max-age=86400, must-revalidate")
		h.ServeHTTP(w, r)
	}
}
