package plug

import (
	"context"
	"net/http"
	"regexp"

	"github.com/gernest/vince/params"
)

func Browser(ctx context.Context) Pipeline {
	return Pipeline{
		Firewall(ctx),
		FetchSession,
		FetchFlash,
		PutSecureBrowserHeaders,
		SessionTimeout,
		Auth,
		LastSeen,
	}
}

func SharedLink() Pipeline {
	return Pipeline{
		PutSecureBrowserHeaders,
	}
}

func Protect() Pipeline {
	return Pipeline{
		CSRF,
		Captcha,
	}
}

func API(ctx context.Context) Pipeline {
	return Pipeline{
		Firewall(ctx),
		FetchSession,
		Auth,
	}
}

func InternalStatsAPI(ctx context.Context) Pipeline {
	return Pipeline{
		Firewall(ctx),
		FetchSession,
		AuthorizedSiteAccess(),
	}
}

func (p Pipeline) Re(exp string, method string, f func(w http.ResponseWriter, r *http.Request)) Plug {
	re := regexp.MustCompile(exp)
	pipe := p.Pass(f)
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if method == r.Method && re.MatchString(r.URL.Path) {
				r = r.WithContext(params.Set(r.Context(),
					params.Re(re, r.URL.Path),
				))
				pipe.ServeHTTP(w, r)
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}

func (p Pipeline) GET(re string, f func(w http.ResponseWriter, r *http.Request)) Plug {
	return p.Re(re, http.MethodGet, f)
}

func (p Pipeline) PUT(re string, f func(w http.ResponseWriter, r *http.Request)) Plug {
	return p.Re(re, http.MethodPut, f)
}

func (p Pipeline) POST(re string, f func(w http.ResponseWriter, r *http.Request)) Plug {
	return p.Re(re, http.MethodPost, f)
}

func (p Pipeline) DELETE(re string, f func(w http.ResponseWriter, r *http.Request)) Plug {
	return p.Re(re, http.MethodDelete, f)
}

func (p Pipeline) PathGET(path string, handler func(w http.ResponseWriter, r *http.Request)) Plug {
	return p.Path(http.MethodGet, path, handler)
}

func (p Pipeline) PathPOST(path string, handler func(w http.ResponseWriter, r *http.Request)) Plug {
	return p.Path(http.MethodPost, path, handler)
}

func (p Pipeline) PathPUT(path string, handler func(w http.ResponseWriter, r *http.Request)) Plug {
	return p.Path(http.MethodPut, path, handler)
}

func (p Pipeline) PathDELETE(path string, handler func(w http.ResponseWriter, r *http.Request)) Plug {
	return p.Path(http.MethodDelete, path, handler)
}

func (p Pipeline) Path(method, path string, handler func(w http.ResponseWriter, r *http.Request)) Plug {
	pipe := p.Pass(handler)
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == method && r.URL.Path == path {
				pipe.ServeHTTP(w, r)
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}
