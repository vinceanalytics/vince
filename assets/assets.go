package assets

import (
	"net/http"
	"strings"

	"github.com/NYTimes/gziphandler"
	"github.com/gernest/vince/assets/tracker"
)

func Match(path string) bool {
	return path == "robots.txt" || path == "favicon.ico" ||
		strings.HasPrefix(path, "/css") ||
		strings.HasPrefix(path, "/js")
}

func Serve() http.Handler {
	h := gziphandler.GzipHandler(serve())
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}

func serve() http.Handler {
	track := tracker.Serve()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if tracker.Match(r.URL.Path) {
			track.ServeHTTP(w, r)
			return
		}
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})
}
