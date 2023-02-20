package router

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"github.com/gernest/vince/api"
	"github.com/gernest/vince/auth"
	"github.com/gernest/vince/health"
	"github.com/gernest/vince/params"
	"github.com/gernest/vince/plug"
	"github.com/gernest/vince/render"
	"github.com/gernest/vince/stats"
)

func Plug(ctx context.Context) plug.Plug {
	pipe := plug.Pipeline{
		API(ctx),
		APIStatsV1(ctx),
		APISitesV1(ctx),
		APIStats(ctx),
		AdminScope(ctx),
	}
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			pipe.Pass(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				h.ServeHTTP(w, r)
			})).ServeHTTP(w, r)
		})
	}
}

func AdminScope(ctx context.Context) plug.Plug {
	pipe := append(plug.Browser(ctx), plug.Protect()...)

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/":
				pipe.Pass(Home).ServeHTTP(w, r)
				return
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

func Home(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusFound)
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

func APIStats(ctx context.Context) plug.Plug {
	pipe := plug.InternalStatsAPI(ctx)
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

var domainStatus = regexp.MustCompile(`^/api/(?P<domain>[^.]+)/status$`)

func API(ctx context.Context) plug.Plug {
	pipe := plug.API(ctx)
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api") {
				switch r.Method {
				case http.MethodGet:
					switch r.URL.Path {
					case "/api/event":
						pipe.Pass(api.Events).ServeHTTP(w, r)
						return
					case "/api/error":
						pipe.Pass(S501).ServeHTTP(w, r)
						return
					case "/api/health":
						pipe.Pass(health.Handle).ServeHTTP(w, r)
						return
					case "/api/system":
						pipe.Pass(api.Info).ServeHTTP(w, r)
						return
					case "/api/sites":
						pipe.Pass(S501).ServeHTTP(w, r)
						return
					default:
						if domainStatus.MatchString(r.URL.Path) {
							matches := domainStatus.FindStringSubmatch(r.URL.Path)
							r = r.WithContext(params.Set(r.Context(), params.Params{
								"domain": matches[domainStatus.SubexpIndex("domain")],
							}))
							pipe.Pass(S501).ServeHTTP(w, r)
							return
						}
					}
				case http.MethodPost:
					switch r.URL.Path {
					case "/subscription/webhook":
						pipe.Pass(S501).ServeHTTP(w, r)
						return
					}
				}
			}
			h.ServeHTTP(w, r)
		})
	}
}

func APIStatsV1(ctx context.Context) plug.Plug {
	pipe := append(plug.PublicAPI(ctx), plug.AuthorizeStatsAPI)
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/v1/stats") {
				if r.Method == http.MethodGet {
					switch r.URL.Path {
					case "/api/v1/stats/realtime/visitors":
						pipe.Pass(stats.V1RealtimeVisitors).ServeHTTP(w, r)
						return
					case "/api/v1/stats/aggregate":
						pipe.Pass(stats.V1Aggregate).ServeHTTP(w, r)
						return
					case "/api/v1/stats/breakdown":
						pipe.Pass(stats.V1Breakdown).ServeHTTP(w, r)
						return
					case "/api/v1/stats/timeseries":
						pipe.Pass(stats.V1Timeseries).ServeHTTP(w, r)
						return
					}
				}
				h.ServeHTTP(w, r)
			}
		})
	}
}

var sites = regexp.MustCompile(`^/api/v1/sites/(?P<site_id>[^.]+)$`)
var goals = regexp.MustCompile(`^/api/v1/sites/goals/(?P<goal_id>[^.]+)$`)

func APISitesV1(ctx context.Context) plug.Plug {
	pipe := append(plug.PublicAPI(ctx), plug.AuthorizeSiteAPI)
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/v1/sites") {
				switch r.URL.Path {
				case "/api/v1/sites":
					if r.Method == http.MethodPost {
						pipe.Pass(S501).ServeHTTP(w, r)
						return
					}
				case "/api/v1/sites/shared-links":
					if r.Method == http.MethodPut {
						pipe.Pass(S501).ServeHTTP(w, r)
						return
					}
				case "/api/v1/sites/goals":
					if r.Method == http.MethodPut {
						pipe.Pass(S501).ServeHTTP(w, r)
						return
					}
				default:
					if goals.MatchString(r.URL.Path) {
						if r.Method == http.MethodDelete {
							pipe.Pass(S501).ServeHTTP(w, r)
							return
						}
					}
					if sites.MatchString(r.URL.Path) {
						switch r.Method {
						case http.MethodGet, http.MethodDelete:
							matches := sites.FindStringSubmatch(r.URL.Path)
							r = r.WithContext(params.Set(r.Context(), params.Params{
								"site_id": matches[sites.SubexpIndex("site_id")],
							}))
							pipe.Pass(S501).ServeHTTP(w, r)
							return
						}
						return
					}
				}
				h.ServeHTTP(w, r)
			}
		})
	}
}
