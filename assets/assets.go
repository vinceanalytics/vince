package assets

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gernest/vince/assets/ui"
	"github.com/gernest/vince/plug"
)

var files = map[string]bool{
	"/android-chrome-192x192.png": true,
	"/favicon-32x32.png":          true,
	"/android-chrome-512x512.png": true,
	"/favicon.ico":                true,
	"/apple-touch-icon.png":       true,
	"/site.webmanifest":           true,
	"/favicon-16x16.png":          true,
	"robots.txt":                  true,
}

func match(path string) bool {
	return strings.HasPrefix(path, "/css") ||
		strings.HasPrefix(path, "/js") ||
		strings.HasPrefix(path, "/fonts") || files[path]
}

func Plug() plug.Plug {
	files, _ := fs.Sub(ui.UI, "app")
	app := http.FileServer(http.FS(files))
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
