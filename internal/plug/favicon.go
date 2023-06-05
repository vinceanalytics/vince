package plug

import (
	_ "embed"
	"net/http"
	"net/url"
	"regexp"

	"github.com/vinceanalytics/vince/internal/referrer"
)

//go:embed placeholder_favicon.ico
var favicon []byte

var source = regexp.MustCompile(`^/favicon/sources/(?P<source>[^.]+)$`)
var DefaultClient = &http.Client{}

type Client interface {
	Do(*http.Request) (*http.Response, error)
}

func Favicon(klient Client) Plug {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/favicon/sources/placeholder" {
				placeholder(w)
				return
			}
			if source.MatchString(r.URL.Path) {
				matches := source.FindStringSubmatch(r.URL.Path)
				src := matches[source.SubexpIndex("source")]
				src, _ = url.QueryUnescape(src)
				if !referrer.ServeFavicon(src, w, r) {
					placeholder(w)
				}
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}

func placeholder(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "image/x-icon")
	w.Header().Set("Cache-Control", "public, max-age=2592000")
	w.WriteHeader(http.StatusOK)
	w.Write(favicon)
}
