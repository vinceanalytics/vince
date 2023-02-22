package plug

import (
	"errors"
	"net/http"

	"github.com/gernest/vince/config"
	"github.com/gernest/vince/log"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/render"
	"github.com/gernest/vince/sites"
	"gorm.io/gorm"
)

func AuthorizedSiteAccess(allowed ...string) Plug {
	if allowed == nil {
		allowed = []string{
			"public", "viewer", "admin", "super_admin", "owner",
		}
	}
	type AuthSite struct {
		ID     uint64
		Public bool
	}
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			db := models.Get(ctx)
			query := r.URL.Query()
			var site AuthSite
			domain := query.Get("domain")
			if domain == "" {
				domain = query.Get("website")
			}
			err := db.Model(&models.Site{}).Where("domain=?", domain).
				Select("id", "public").Limit(1).Find(&site).Error
			if err != nil {
				log.Get(ctx).Err(err).Str("domain", domain).Msg("failed to get site by domain")
				render.ERROR(ctx, w, http.StatusNotFound)
				return
			}
			var sharedLink uint64
			slug := r.URL.Query().Get("auth")
			if slug != "" {
				err = db.Model(&models.SharedLink{}).Where("slug=?", slug).
					Select("site_id").Limit(1).Find(&sharedLink).Error
				if err != nil {
					if !errors.Is(err, gorm.ErrRecordNotFound) {
						log.Get(ctx).Err(err).Str("slug", slug).Msg("failed to get shared link")
						render.ERROR(ctx, w, http.StatusInternalServerError)
						return
					}
				}
			}
			usr := models.GetCurrentUser(ctx)

			var membership string
			if usr != nil {
				membership = sites.Role(ctx, usr.ID, site.ID)
			}

			var role string
			switch {
			case membership != "":
				role = membership
			case usr != nil && config.Get(ctx).IsSuperUser(usr.ID):
				role = "super_admin"
			case site.Public:
				role = "public"
			case sharedLink == site.ID:
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
