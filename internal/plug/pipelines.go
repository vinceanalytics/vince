package plug

import (
	"net/http"
	"strings"
)

func Browser() Pipeline {
	return Pipeline{
		PutSecureBrowserHeaders,
	}
}

func API() Pipeline {
	return Pipeline{
		AcceptJSON,
	}
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
			if (method == r.Method) && r.URL.Path == path {
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

func Ok(ok bool, pipe Plug) Plug {
	if ok {
		return pipe
	}
	return func(h http.Handler) http.Handler {
		return h
	}
}

func bearer(h http.Header) string {
	a := h.Get("authorization")
	if a == "" {
		return ""
	}
	if !strings.HasPrefix(a, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(a, "Bearer "))
}
