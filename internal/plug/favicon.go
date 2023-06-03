package plug

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"

	"github.com/vinceanalytics/vince/internal/referrer"
	"github.com/vinceanalytics/vince/pkg/log"
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
				domain := referrer.Favicon(src)
				if domain == "" {
					placeholder(w)
					return
				}
				req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("https://icons.duckduckgo.com/ip3/%s.ico", domain), nil)
				res, err := klient.Do(req)
				if err != nil {
					log.Get().Err(err).Str("domain", domain).
						Msg("failed getting icon from duckduckgo")
					placeholder(w)
					return
				}
				defer res.Body.Close()
				if res.StatusCode == http.StatusOK {
					var b bytes.Buffer
					io.Copy(&b, res.Body)
					if !bytes.Contains(b.Bytes(), []byte{137, 80, 78, 71, 13, 10, 26, 10}) {
						// set forward headers
						h := []string{"content-type", "cache-control", "expires"}
						for _, v := range h {
							if x := res.Header.Get(v); x != "" {
								w.Header().Set(v, x)
							}
						}
						if bytes.HasPrefix(b.Bytes(), []byte("<svg")) {
							w.Header().Set("content-type", "image/svg+xml")
						} else {
							w.Header().Set("content-type", res.Header.Get("content-type"))
						}
						w.Header().Set("content-security-policy", "script-src 'none'")
						w.Header().Set("content-disposition", "attachment")
						w.WriteHeader(http.StatusOK)
						w.Write(b.Bytes())
						return
					}

				}
				placeholder(w)
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
