package assets

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gernest/vince/assets/ui"
	"github.com/gernest/vince/plug"
)

func match(path string) bool {
	return path == "robots.txt" || path == "favicon.ico" ||
		strings.HasPrefix(path, "/css") ||
		strings.HasPrefix(path, "/js")
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
