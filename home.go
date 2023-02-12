package vince

import (
	"net/http"

	"github.com/gernest/vince/sessions"
)

func (v *Vince) home() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ss, r := sessions.Load(r)
		ss.Data["login_dest"] = r.URL.Path
		ss.Save(w)
		http.Redirect(w, r, "/login", http.StatusFound)
	})
}
