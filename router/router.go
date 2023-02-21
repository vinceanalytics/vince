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
		Share(),
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
	pipeAccount := append(pipe, plug.RequireAccount)
	expr := plug.ExprPipe{
		pipe.GET(`^/register/invitation/(?P<invitation_id>[^.]+)$`, S501),
		pipe.POST(`^/register/invitation/(?P<invitation_id>[^.]+)$`, S501),
		pipe.DELETE(`^/settings/api-keys/(?P<id>[^.]+)$`, S501),
		pipe.GET(`^/billing/change-plan/preview/(?P<plan_id>[^.]+)$`, S501),
		pipe.POST(`^/billing/change-plan/(?P<new_plan_id>[^.]+)$`, S501),
	}
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/":
				pipe.Pass(Home).ServeHTTP(w, r)
				return
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
					pipeAccount.Pass(auth.ActivateForm).ServeHTTP(w, r)
					return
				}
				if r.Method == http.MethodPost {
					pipe.Pass(auth.Activate).ServeHTTP(w, r)
					return
				}
			case "/activate/request-code":
				pipe.Pass(S501).ServeHTTP(w, r)
				return
			case "/login":
				if r.Method == http.MethodGet {
					pipe.Pass(auth.LoginForm).ServeHTTP(w, r)
					return
				}
			case "/password/request-reset":
				if r.Method == http.MethodGet {
					pipe.Pass(S501).ServeHTTP(w, r)
					return
				}
				if r.Method == http.MethodPost {
					pipe.Pass(S501).ServeHTTP(w, r)
					return
				}
			case "/password/reset":
				if r.Method == http.MethodGet {
					pipe.Pass(S501).ServeHTTP(w, r)
					return
				}
				if r.Method == http.MethodPost {
					pipe.Pass(S501).ServeHTTP(w, r)
					return
				}
			case "/error_report":
				if r.Method == http.MethodPost {
					pipe.Pass(S501).ServeHTTP(w, r)
					return
				}
			case "/password":
				if r.Method == http.MethodGet {
					pipe.Pass(S501).ServeHTTP(w, r)
					return
				}
				if r.Method == http.MethodPost {
					pipe.Pass(S501).ServeHTTP(w, r)
					return
				}
			case "/logout":
				if r.Method == http.MethodGet {
					pipe.Pass(S501).ServeHTTP(w, r)
					return
				}
			case "/settings":
				if r.Method == http.MethodGet {
					pipe.Pass(S501).ServeHTTP(w, r)
					return
				}
				if r.Method == http.MethodPut {
					pipe.Pass(S501).ServeHTTP(w, r)
					return
				}
			case "/me":
				if r.Method == http.MethodDelete {
					pipe.Pass(S501).ServeHTTP(w, r)
					return
				}
			case "/settings/api-keys/new":
				if r.Method == http.MethodGet {
					pipe.Pass(S501).ServeHTTP(w, r)
					return
				}
				if r.Method == http.MethodPost {
					pipe.Pass(S501).ServeHTTP(w, r)
					return
				}
			case "/auth/google/callback":
				if r.Method == http.MethodGet {
					pipe.Pass(S501).ServeHTTP(w, r)
					return
				}
			case "/billing/change-plan":
				if r.Method == http.MethodGet {
					pipe.Pass(S501).ServeHTTP(w, r)
					return
				}
			default:
				if expr.ServeHTTP(w, r) {
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

func APIStats(ctx context.Context) plug.Plug {
	pipe := plug.InternalStatsAPI(ctx)
	expr := plug.ExprPipe{
		pipe.GET(`^/api/stats/(?P<domain>[^.]+)/current-visitors$`, S501),
		pipe.GET(`^/api/stats/(?P<domain>[^.]+)/main-graph$`, S501),
		pipe.GET(`^/api/stats/(?P<domain>[^.]+)/top-stats$`, S501),
		pipe.GET(`^/api/stats/(?P<domain>[^.]+)/sources$`, S501),
		pipe.GET(`^/api/stats/(?P<domain>[^.]+)/utm_mediums$`, S501),
		pipe.GET(`^/api/stats/(?P<domain>[^.]+)/utm_sources$`, S501),
		pipe.GET(`^/api/stats/(?P<domain>[^.]+)/utm_campaigns$`, S501),
		pipe.GET(`^/api/stats/(?P<domain>[^.]+)/utm_contents$`, S501),
		pipe.GET(`^/api/stats/(?P<domain>[^.]+)/utm_terms$`, S501),
		pipe.GET(`^/api/stats/(?P<domain>[^.]+)/referrers/(?P<referrer>[^.]+)$`, S501),
		pipe.GET(`^/api/stats/(?P<domain>[^.]+)/pages$`, S501),
		pipe.GET(`^/api/stats/(?P<domain>[^.]+)/entry-pages$`, S501),
		pipe.GET(`^/api/stats/(?P<domain>[^.]+)/exit-pages$`, S501),
		pipe.GET(`^/api/stats/(?P<domain>[^.]+)/countries$`, S501),
		pipe.GET(`^/api/stats/(?P<domain>[^.]+)/regions$`, S501),
		pipe.GET(`^/api/stats/(?P<domain>[^.]+)/cities$`, S501),
		pipe.GET(`^/api/stats/(?P<domain>[^.]+)/browsers$`, S501),
		pipe.GET(`^/api/stats/(?P<domain>[^.]+)/browser-versions$`, S501),
		pipe.GET(`^/api/stats/(?P<domain>[^.]+)/operating-systems$`, S501),
		pipe.GET(`^/api/stats/(?P<domain>[^.]+)/operating-system-versions$`, S501),
		pipe.GET(`^/api/stats/(?P<domain>[^.]+)/screen-sizes$`, S501),
		pipe.GET(`^/api/stats/(?P<domain>[^.]+)/conversions$`, S501),
		pipe.GET(`^/api/stats/(?P<domain>[^.]+)/property/(?P<property>[^.]+)$`, S501),
		pipe.GET(`^/api/stats/(?P<domain>[^.]+)/suggestions/(?P<filter_name>[^.]+)$`, S501),
	}

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/stats") {
				pipe.Pass(S404).ServeHTTP(w, r)
				if expr.ServeHTTP(w, r) {
					return
				}
			}
			h.ServeHTTP(w, r)
		})
	}
}

func S501(w http.ResponseWriter, r *http.Request) {
	render.ERROR(r.Context(), w, http.StatusNotImplemented)
}

func S404(w http.ResponseWriter, r *http.Request) {
	render.ERROR(r.Context(), w, http.StatusNotFound)
}

func API(ctx context.Context) plug.Plug {
	var domainStatus = regexp.MustCompile(`^/api/(?P<domain>[^.]+)/status$`)
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
							r = r.WithContext(params.Set(r.Context(),
								params.Re(domainStatus, r.URL.Path),
							))
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

func APISitesV1(ctx context.Context) plug.Plug {
	var sites = regexp.MustCompile(`^/api/v1/sites/(?P<site_id>[^.]+)$`)
	var goals = regexp.MustCompile(`^/api/v1/sites/goals/(?P<goal_id>[^.]+)$`)
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
							r = r.WithContext(params.Set(r.Context(),
								params.Re(goals, r.URL.Path),
							))
							pipe.Pass(S501).ServeHTTP(w, r)
							return
						}
					}
					if sites.MatchString(r.URL.Path) {
						switch r.Method {
						case http.MethodGet, http.MethodDelete:
							r = r.WithContext(params.Set(r.Context(),
								params.Re(sites, r.URL.Path),
							))
							pipe.Pass(S501).ServeHTTP(w, r)
							return
						}
					}
				}
				h.ServeHTTP(w, r)
			}
		})
	}
}

func Share() plug.Plug {
	var domain = regexp.MustCompile(`^/share/(?P<domain>[^.]+)$`)
	var auth = regexp.MustCompile(`^/share/(?P<slug>[^.]+)/authenticate$`)
	pipe := plug.SharedLink()
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/share/") {
				if domain.MatchString(r.URL.Path) {
					if r.Method == http.MethodGet {
						r = r.WithContext(params.Set(r.Context(),
							params.Re(domain, r.URL.Path),
						))
						pipe.Pass(S501).ServeHTTP(w, r)
						return
					}
				}
				if auth.MatchString(r.URL.Path) {
					if r.Method == http.MethodGet {
						r = r.WithContext(params.Set(r.Context(),
							params.Re(auth, r.URL.Path),
						))
						pipe.Pass(S501).ServeHTTP(w, r)
						return
					}
				}
			}

			h.ServeHTTP(w, r)
		})
	}
}
