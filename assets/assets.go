package assets

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/internal/plug"
)

var files = map[string]bool{
	"/favicon.svg": true,
	"/favicon.ico": true,
	"/favicon":     true,
	"/robots.txt":  true,
	"/logo.svg":    true,
	"/index.html":  true,
	"/":            true,
}

//go:embed ui
var static embed.FS

var ui = must.Must(fs.Sub(static, "ui"))("failed getting sub directory")

var FS = http.FileServer(http.FS(ui))

func Match(path string) bool {
	return strings.HasPrefix(path, "/static") ||
		strings.HasPrefix(path, "/vs") ||
		strings.HasPrefix(path, "/min-map") ||
		files[path]
}

func Plug() plug.Plug {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if Match(r.URL.Path) {
				FS.ServeHTTP(w, r)
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}
