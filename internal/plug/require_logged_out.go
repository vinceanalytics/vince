package plug

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/sessions"
)

func RequireLoggedOut(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if models.GetUser(r.Context()) != nil {
			session, r := sessions.Load(r)
			session.Data.LoggedIn = true
			session.Save(r.Context(), w)
			http.Redirect(w, r, "/sites", http.StatusFound)
			return
		}
		h.ServeHTTP(w, r)
	})
}