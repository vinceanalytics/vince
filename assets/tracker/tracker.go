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
		h.ServeHTTP(w, r)
	}
}
