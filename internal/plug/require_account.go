package plug

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/params"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/sessions"
)

func RequireAccount(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		usr := models.GetUser(ctx)
		if usr == nil {
			session, r := sessions.Load(r)
			session.Data.LoginDest = r.URL.Path
			session.Save(r.Context(), w)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		if !usr.EmailVerified && r.URL.Path == "/activate/me" {
			http.Redirect(w, r, "/activate", http.StatusFound)
			return
		}

		g := params.Get(ctx)
		if s := g.Get("site"); s != "" {
			// accessing user site
			site := models.SiteByDomain(ctx, s)
			if site == nil {
				render.ERROR(ctx, w, http.StatusNotFound)
				return
			}
			ctx = models.SetSite(ctx, site)
			r = r.WithContext(ctx)
		}
		h.ServeHTTP(w, r)
	})
}
