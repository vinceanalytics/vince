package plug

import (
	"net/http"

	"github.com/gernest/vince/models"
	"github.com/gernest/vince/sessions"
)

func RequireLoggedOut(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if models.GetCurrentUser(r.Context()) != nil {
			session, r := sessions.Load(r)
			session.Data.LoggedIn = true
			session.Save(w)
			http.Redirect(w, r, "/sites", http.StatusFound)
			return
		}
		h.ServeHTTP(w, r)
	})
}
