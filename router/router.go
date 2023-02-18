package router

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gernest/vince/auth"
	"github.com/gernest/vince/params"
	"github.com/gernest/vince/plug"
	"github.com/gernest/vince/render"
)

func Plug() plug.Plug {
	pipe := plug.Pipeline{
		Home(),
		APIStats(),
		AdminScope(),
	}
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			pipe.Pass(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				h.ServeHTTP(w, r)
			})).ServeHTTP(w, r)
		})
	}
}

func AdminScope() plug.Plug {
	pipe := plug.Pipeline{
		plug.FetchSession,
		plug.PutSecureBrowserHeaders,
		plug.SessionTimeout,
		plug.Auth,
		plug.LastSeen,
		plug.Captcha,
		plug.CSRF,
	}
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/login":
				if r.Method == http.MethodGet {
					pipe.Pass(auth.LoginForm).ServeHTTP(w, r)
					return
				}
			case "/register":
				if r.Method == http.MethodGet {
					pipe.Pass(auth.RegisterForm).ServeHTTP(w, r)
					return
				}
				if r.Method == http.MethodPost {
					pipe.Pass(auth.Register).ServeHTTP(w, r)
					return
				}
			case "/activate":
				if r.Method == http.MethodGet {
					pipe.Pass(auth.ActivateForm).ServeHTTP(w, r)
					return
				}
				if r.Method == http.MethodPost {
					pipe.Pass(auth.Activate).ServeHTTP(w, r)
					return
				}
			}
			h.ServeHTTP(w, r)
		})
	}
}

func Home() plug.Plug {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}

var handlesAPIStats = []*ReHandle{
	{
		Re:     regexp.MustCompile(`^/api/stats/(?P<domain>[^.]+)/current-visitors$`),
		Hand:   S501,
		Params: []string{"domain"},
	},
	{
		Re:     regexp.MustCompile(`^/api/stats/(?P<domain>[^.]+)/main-graph$`),
		Hand:   S501,
		Params: []string{"domain"},
	},
	{
		Re:     regexp.MustCompile(`^/api/stats/(?P<domain>[^.]+)/top-stats$`),
		Hand:   S501,
		Params: []string{"domain"},
	},
	{
		Re:     regexp.MustCompile(`^/api/stats/(?P<domain>[^.]+)/sources$`),
		Hand:   S501,
		Params: []string{"domain"},
	},
	{
		Re:     regexp.MustCompile(`^/api/stats/(?P<domain>[^.]+)/utm_mediums$`),
		Hand:   S501,
		Params: []string{"domain"},
	},
	{
		Re:     regexp.MustCompile(`^/api/stats/(?P<domain>[^.]+)/utm_sources$`),
		Hand:   S501,
		Params: []string{"domain"},
	},
	{
		Re:     regexp.MustCompile(`^/api/stats/(?P<domain>[^.]+)/utm_campaigns$`),
		Hand:   S501,
		Params: []string{"domain"},
	},
	{
		Re:     regexp.MustCompile(`^/api/stats/(?P<domain>[^.]+)/utm_contents$`),
		Hand:   S501,
		Params: []string{"domain"},
	},
	{
		Re:     regexp.MustCompile(`^/api/stats/(?P<domain>[^.]+)/utm_terms$`),
		Hand:   S501,
		Params: []string{"domain"},
	},
	{
		Re:     regexp.MustCompile(`^/api/stats/(?P<domain>[^.]+)/referrers/(?P<referrer>[^.]+)$`),
		Hand:   S501,
		Params: []string{"domain", "referrer"},
	},
	{
		Re:     regexp.MustCompile(`^/api/stats/(?P<domain>[^.]+)/pages$`),
		Hand:   S501,
		Params: []string{"domain"},
	},
	{
		Re:     regexp.MustCompile(`^/api/stats/(?P<domain>[^.]+)/entry-pages$`),
		Hand:   S501,
		Params: []string{"domain"},
	},
	{
		Re:     regexp.MustCompile(`^/api/stats/(?P<domain>[^.]+)/exit-pages$`),
		Hand:   S501,
		Params: []string{"domain"},
	},
	{
		Re:     regexp.MustCompile(`^/api/stats/(?P<domain>[^.]+)/countries$`),
		Hand:   S501,
		Params: []string{"domain"},
	},
	{
		Re:     regexp.MustCompile(`^/api/stats/(?P<domain>[^.]+)/regions$`),
		Hand:   S501,
		Params: []string{"domain"},
	},
	{
		Re:     regexp.MustCompile(`^/api/stats/(?P<domain>[^.]+)/cities$`),
		Hand:   S501,
		Params: []string{"domain"},
	},
	{
		Re:     regexp.MustCompile(`^/api/stats/(?P<domain>[^.]+)/browsers$`),
		Hand:   S501,
		Params: []string{"domain"},
	},
	{
		Re:     regexp.MustCompile(`^/api/stats/(?P<domain>[^.]+)/browser-versions$`),
		Hand:   S501,
		Params: []string{"domain"},
	},
	{
		Re:     regexp.MustCompile(`^/api/stats/(?P<domain>[^.]+)/operating-systems$`),
		Hand:   S501,
		Params: []string{"domain"},
	},
	{
		Re:     regexp.MustCompile(`^/api/stats/(?P<domain>[^.]+)/operating-system-versions$`),
		Hand:   S501,
		Params: []string{"domain"},
	},
	{
		Re:     regexp.MustCompile(`^/api/stats/(?P<domain>[^.]+)/screen-sizes$`),
		Hand:   S501,
		Params: []string{"domain"},
	},
	{
		Re:     regexp.MustCompile(`^/api/stats/(?P<domain>[^.]+)/conversions$`),
		Hand:   S501,
		Params: []string{"domain"},
	},
	{
		Re:     regexp.MustCompile(`^/api/stats/(?P<domain>[^.]+)/property/(?P<property>[^.]+)$`),
		Hand:   S501,
		Params: []string{"domain", "property"},
	},
	{
		Re:     regexp.MustCompile(`^/api/stats/(?P<domain>[^.]+)/suggestions/(?P<filter_name>[^.]+)$`),
		Hand:   S501,
		Params: []string{"domain", "filter_name"},
	},
}

func APIStats() plug.Plug {
	pipe := plug.Pipeline{}
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/stats") {
				if r.Method == http.MethodGet {
					for _, handle := range handlesAPIStats {
						if handle.Re.MatchString(r.URL.Path) {
							matches := handle.Re.FindStringSubmatch(r.URL.Path)
							m := make(params.Params)
							for _, p := range handle.Params {
								m[p] = matches[handle.Re.SubexpIndex(p)]
							}
							r = r.WithContext(params.Set(r.Context(), m))
							pipe.Pass(handle.Hand).ServeHTTP(w, r)
							return
						}
					}
				}
				pipe.Pass(S404).ServeHTTP(w, r)
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}

type ReHandle struct {
	Re     *regexp.Regexp
	Hand   http.HandlerFunc
	Params []string
}

func S501(w http.ResponseWriter, r *http.Request) {
	render.ERROR(r.Context(), w, http.StatusNotImplemented)
}

func S404(w http.ResponseWriter, r *http.Request) {
	render.ERROR(r.Context(), w, http.StatusNotFound)
}

func NOOP(w http.ResponseWriter, r *http.Request) {}
