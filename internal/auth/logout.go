package auth

import (
	"net/http"

	"github.com/vinceanalytics/vince/sessions"
)

func Logout(w http.ResponseWriter, r *http.Request) {
	redirect := r.URL.Query().Get("redirect")
	if redirect == "" {
		redirect = "/"
	}
	session, r := sessions.Load(r)
	session.Data = sessions.Data{}
	session.Save(r.Context(), w)
	http.Redirect(w, r, redirect, http.StatusFound)
}
