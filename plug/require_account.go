package plug

import (
	"net/http"

	"github.com/gernest/vince/models"
	"github.com/gernest/vince/sessions"
)

func RequireAccount(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		usr := models.GetCurrentUser(r.Context())
		if usr == nil {
			session, r := sessions.Load(r)
			session.Data.LoginDest = r.URL.Path
			session.Save(w)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		if !usr.EmailVerified && r.URL.Path == "/activate/me" {
			http.Redirect(w, r, "/activate", http.StatusFound)
			return
		}
		h.ServeHTTP(w, r)
	})
}
