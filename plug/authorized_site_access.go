package plug

import (
	"net/http"

	"github.com/gernest/vince/config"
	"github.com/gernest/vince/log"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/render"
)

func AuthorizedSiteAccess(allowed ...string) Plug {
	if allowed == nil {
		allowed = []string{
			"public", "viewer", "admin", "super_admin", "owner",
		}
	}
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			user := models.GetCurrentUser(ctx)
			db := models.Get(ctx)
			query := r.URL.Query()
			var site models.Site
			domain := query.Get("domain")
			if domain == "" {
				domain = query.Get("website")
			}
			var userID uint64
			if user != nil {
				userID = user.ID
			}

			err := db.Where("domain=?", domain).
				Preload("SharedLinks", "slug =?", query.Get("auth")).
				Preload("SiteMemberships", "user_id=?", userID).First(&site).Error
			if err != nil {
				log.Get(ctx).Err(err).Str("domain", domain).Msg("failed to get site by domain")
				render.ERROR(ctx, w, http.StatusNotFound)
				return
			}
			var role string
			switch {
			case len(site.SiteMemberships) == 1:
				role = site.SiteMemberships[0].Role
			case user != nil && config.Get(ctx).IsSuperUser(user.ID):
				role = "super_admin"
			case site.Public:
				role = "public"
			case len(site.SharedLinks) == 1:
				role = "public"
			}
			for _, a := range allowed {
				if a == role {
					r = r.WithContext(models.SetCurrentUserRole(ctx, role))
					h.ServeHTTP(w, r)
					return
				}
			}
			render.ERROR(ctx, w, http.StatusNotFound)
		})
	}
}
