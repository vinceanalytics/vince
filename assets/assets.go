package assets

import (
	"embed"
	"net/http"
	"strings"

	"github.com/vinceanalytics/vince/internal/plug"
)

//go:generate go run gen/main.go app ../../assets/

var files = map[string]bool{
	"/favicon.svg": true,
	"/favicon.ico": true,
	"/favicon":     true,
	"/robots.txt":  true,
}

//go:embed css image js *.svg
var fs embed.FS

func match(path string) bool {
	return strings.HasPrefix(path, "/css") ||
		strings.HasPrefix(path, "/js") ||
		strings.HasPrefix(path, "/fonts") ||
		strings.HasPrefix(path, "/image") || files[path]
}

func Plug() plug.Plug {
	app := http.FileServer(http.FS(fs))
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if match(r.URL.Path) {
				app.ServeHTTP(w, r)
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}
