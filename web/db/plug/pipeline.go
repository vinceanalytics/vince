package plug

import (
	"net/http"
	"strings"

	"github.com/gernest/len64/web/db"
)

type Handler func(db *db.Config, w http.ResponseWriter, r *http.Request)

type Middleware func(h Handler) Handler

type Pipeline []Middleware

func (p Pipeline) With(m ...Middleware) Pipeline {
	return append(p, m...)
}

func (p Pipeline) Then(h Handler) func(db *db.Config, w http.ResponseWriter, r *http.Request) {
	for i := range p {
		h = p[len(p)-1-i](h)
	}
	return h
}

func Browser() Pipeline {
	return Pipeline{
		FetchSession,
		FetchFlash,
		SecureHeaders,
		SessionTimeout,
		FetchFlash,
	}
}

func FetchSession(h Handler) Handler {
	return func(db *db.Config, w http.ResponseWriter, r *http.Request) {
		db.Load(w, r)
		h(db, w, r)
	}
}

func FetchFlash(h Handler) Handler {
	return func(db *db.Config, w http.ResponseWriter, r *http.Request) {
		db.Flash(w)
		h(db, w, r)
	}
}

func Track(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/public/js/len64") {
			w.Header().Set("x-content-type-options", "nosniff")
			w.Header().Set("cross-origin-resource-policy", "cross-origin")
			w.Header().Set("access-control-allow-origin", "*")
			w.Header().Set("cache-control", "public, max-age=86400, must-revalidate")
		}
		h.ServeHTTP(w, r)
	})
}

func SessionTimeout(h Handler) Handler {
	return func(db *db.Config, w http.ResponseWriter, r *http.Request) {
		db.SessionTimeout(w)
		h(db, w, r)
	}
}

func SecureHeaders(h Handler) Handler {
	return func(db *db.Config, w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-frame-options", "SAMEORIGIN")
		w.Header().Set("x-xss-protection", "1; mode=block")
		w.Header().Set("x-content-type-options", "nosniff")
		w.Header().Set("x-download-options", "noopen")
		w.Header().Set("x-permitted-cross-domain-policies", "none")
		w.Header().Set("cross-origin-window-policy", "deny")
		h(db, w, r)
	}
}

func Captcha(h Handler) Handler {
	return func(db *db.Config, w http.ResponseWriter, r *http.Request) {
		db.SaveCaptcha(w)
		h(db, w, r)
	}
}

func CSRF(h Handler) Handler {
	return func(db *db.Config, w http.ResponseWriter, r *http.Request) {
		db.SaveCsrf(w)
		h(db, w, r)
	}
}

func VerifyCSRF(h Handler) Handler {
	return func(db *db.Config, w http.ResponseWriter, r *http.Request) {
		if !db.IsValidCsrf(r) {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		h(db, w, r)
	}
}

func RequireAccount(h Handler) Handler {
	return func(db *db.Config, w http.ResponseWriter, r *http.Request) {
		if !db.Authorize(w, r) {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		h(db, w, r)
	}
}

func RequireLogout(h Handler) Handler {
	return func(db *db.Config, w http.ResponseWriter, r *http.Request) {
		if !db.Logout(w) {
			http.Redirect(w, r, "/sites", http.StatusFound)
			return
		}
		h(db, w, r)
	}
}
