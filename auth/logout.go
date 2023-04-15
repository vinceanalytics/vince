package auth

import (
	"net/http"

	"github.com/gernest/vince/sessions"
)

func Logout(w http.ResponseWriter, r *http.Request) {
	redirect := r.URL.Query().Get("redirect")
	if redirect == "" {
		redirect = "/"
	}
	session, r := sessions.Load(r)
	session.Data = sessions.Data{}
	session.Save(w)
	http.Redirect(w, r, redirect, http.StatusFound)
}
