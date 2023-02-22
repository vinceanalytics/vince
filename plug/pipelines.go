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

func PublicAPI(ctx context.Context) Pipeline {
	return Pipeline{
		Firewall(ctx),
	}
}

type Expr interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request) bool
}

type ExprPipe []Expr

func (ex ExprPipe) ServeHTTP(w http.ResponseWriter, r *http.Request) bool {
	for _, e := range ex {
		if e.ServeHTTP(w, r) {
			return true
		}
	}
	return false
}

type ExprFunc func(w http.ResponseWriter, r *http.Request) bool

func (f ExprFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) bool {
	return f(w, r)
}

func ExprHandle(h http.Handler, expr ...Expr) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, e := range expr {
			if e.ServeHTTP(w, r) {
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

func Re(exp string, method string, f http.Handler) Expr {
	re := regexp.MustCompile(exp)
	return ExprFunc(func(w http.ResponseWriter, r *http.Request) bool {
		if method == r.Method && re.MatchString(r.URL.Path) {
			r = r.WithContext(params.Set(r.Context(),
				params.Re(re, r.URL.Path),
			))
			f.ServeHTTP(w, r)
			return true
		}
		return false
	})
}

func (p Pipeline) GET(re string, f http.HandlerFunc) Expr {
	return Re(re, http.MethodGet, p.Pass(f))
}

func (p Pipeline) PUT(re string, f http.HandlerFunc) Expr {
	return Re(re, http.MethodPut, p.Pass(f))
}

func (p Pipeline) POST(re string, f http.HandlerFunc) Expr {
	return Re(re, http.MethodPost, p.Pass(f))
}

func (p Pipeline) DELETE(re string, f http.HandlerFunc) Expr {
	return Re(re, http.MethodDelete, p.Pass(f))
}
