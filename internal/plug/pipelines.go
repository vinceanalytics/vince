package plug

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/dlclark/regexp2"
	"github.com/vinceanalytics/vince/internal/params"
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
	}
}

func (p Pipeline) Re(exp string, method string, f func(w http.ResponseWriter, r *http.Request)) Plug {
	for k, v := range replace {
		exp = strings.ReplaceAll(exp, k, v)
	}
	re := regexp2.MustCompile(exp, regexp2.ECMAScript)
	pipe := p.Pass(f)
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if checkMethod(method, r) {
				if m, err := re.FindStringMatch(r.URL.Path); err == nil && m != nil {
					r = r.WithContext(params.Set(r.Context(),
						params.Re(m),
					))
					pipe.ServeHTTP(w, r)
					return
				}
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
			if checkMethod(method, r) && r.URL.Path == path {
				pipe.ServeHTTP(w, r)
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}

func (p Pipeline) Prefix(path string, handler func(w http.ResponseWriter, r *http.Request)) Plug {
	pipe := p.Pass(handler)
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, path) {
				pipe.ServeHTTP(w, r)
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}

func PREFIX(prefix string, pipe ...Plug) Plug {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, prefix) {
				Pipeline(pipe).Pass(NOOP).ServeHTTP(w, r)
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}

const (
	// User nickname regex
	// may only contain alphanumeric characters or hyphens.
	// cannot have multiple consecutive hyphens.
	// cannot begin or end with a hyphen.
	// Maximum is 39 characters.
	owner = `(?<owner>[a-z\d](?:[a-z\d]|-(?=[a-z\d])){0,38})`
	site  = `(?<site>\b((?=[a-z0-9-]{1,63}\.)(xn--)?[a-z0-9]+(-[a-z0-9]+)*\.)+[a-z]{2,63}\b)`
	id    = `(?<%s>\d*)`
)

func reid(x string) string {
	return fmt.Sprintf(id, x)
}

var replace = map[string]string{
	":owner": owner,
	":site":  site,
	":goal":  reid("goal"),
}

func Ok(ok bool, pipe Plug) Plug {
	if ok {
		return pipe
	}
	return func(h http.Handler) http.Handler {
		return h
	}
}

const form = "application/x-www-form-urlencoded"

func checkMethod(method string, r *http.Request) bool {
	if method == r.Method {
		return true
	}
	if r.Header.Get("content-type") == form {
		return r.FormValue("_method") == method
	}
	return false
}
