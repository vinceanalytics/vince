package vince

import (
	"net/http"
	"regexp"
)

var registerInvitationRe = regexp.MustCompile(`^/register/invitation/(?P<v0>[^.]+)$`)

func isAdminPath(path, method string) bool {
	switch method {
	case http.MethodGet:
		switch path {
		case "/register", "/activate", "/login",
			"/password/request-reset", "/password/reset":
			return true
		default:
			if registerInvitationRe.MatchString(path) {
				return true
			}
			return false
		}
	case http.MethodPost:
		switch path {
		case "/register", "/activate/request-code",
			"/activate", "/login", "/password/request-reset",
			"/password/reset":
			return true
		default:
			if registerInvitationRe.MatchString(path) {
				return true
			}
			return false
		}
	default:
		return false
	}
}

func (v *Vince) admin() http.Handler {
	loginForm := v.loginForm()
	registerForm := v.registerForm()
	register := v.register()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login":
			if r.Method == http.MethodGet {
				loginForm.ServeHTTP(w, r)
				return
			}
		case "/register":
			if r.Method == http.MethodGet {
				registerForm.ServeHTTP(w, r)
				return
			}
			if r.Method == http.MethodPost {
				register.ServeHTTP(w, r)
				return
			}
		case "/activate":
			if r.Method == http.MethodGet {
				ActivateForm(w, r)
				return
			}
		}
		ServeError(r.Context(), w, http.StatusNotImplemented)
	})
}
