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
	login := v.login()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login":
			if r.Method == http.MethodGet {
				login.ServeHTTP(w, r)
				return
			}
		}
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
	})
}
