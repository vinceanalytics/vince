package tracker

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed js/*.js
var files embed.FS

func Match(s string) bool {
	return strings.HasPrefix(s, "/js/vince")
}

func Serve() http.HandlerFunc {
	sub, _ := fs.Sub(files, "js")
	h := http.StripPrefix("/js", http.FileServer(http.FS(sub)))
	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}
