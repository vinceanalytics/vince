package vince

import "net/http"

func (v *Vince) home() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ss, r := v.clientSession.Load(r)
		ss.Data["login_dest"] = r.URL.Path
		_ = ss.Save(w)
		http.Redirect(w, r, "/login", http.StatusFound)
	})
}
