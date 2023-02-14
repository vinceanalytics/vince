package server

import (
	"net/http"

	"github.com/gernest/vince/api"
	"github.com/gernest/vince/health"
)

func (v *Vince) api(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		switch r.URL.Path {
		case "/api/event":
			api.Events(w, r)
			return
		case "/subscription/webhook":
			http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
			return
		}
	case http.MethodGet:
		switch r.URL.Path {
		case "/api/error":
			http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
			return
		case "/api/health":
			health.Handle(w, r)
			return
		case "/api/system":
			api.Info(w, r)
			return
		default:
			if domainStatusRe.MatchString(r.URL.Path) {
				domain := domainStatusRe.FindStringSubmatch(r.URL.Path)[1]
				_ = domain
				http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
				return
			}
		}
	}
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}
