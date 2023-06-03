package plug

import (
	"context"
	"net/http"

	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/params"
	"github.com/vinceanalytics/vince/render"
	"github.com/vinceanalytics/vince/sites"
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
			params := params.Get(r.Context())
			domain := params["domain"]
			if domain == "" {
				domain = params["website"]
			}
			site := GetSite(ctx, domain)
			if site == nil {
				render.ERROR(ctx, w, http.StatusNotFound)
				return
			}
			slug := r.URL.Query().Get("auth")
			sharedLink := GetSharedLink(ctx, slug)
			usr := models.GetUser(ctx)

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
			case sharedLink != nil && sharedLink.SiteID == site.ID:
				role = "public"
			}
			for _, a := range allowed {
				if a == role {
					ctx = models.SetRole(ctx, role)
					ctx = models.SetSite(ctx, site)
					r = r.WithContext(ctx)
					h.ServeHTTP(w, r)
					return
				}
			}
			render.ERROR(ctx, w, http.StatusNotFound)
		})
	}
}

func GetSite(ctx context.Context, domain string) *models.Site {
	var site models.Site
	err := models.Get(ctx).Model(&models.Site{}).Where("domain=?", domain).
		First(&site).Error
	if err != nil {
		models.LOG(ctx, err, "failed to get site by domain")
		return nil
	}
	return &site
}

func GetSharedLink(ctx context.Context, slug string) *models.SharedLink {
	if slug == "" {
		return nil
	}
	var link models.SharedLink
	err := models.Get(ctx).
		Model(&models.SharedLink{}).
		Where("slug = ?", slug).
		Select("site_id").Limit(1).First(&link).Error

	if err != nil {
		models.LOG(ctx, err, "failed to get shared link")
		return nil
	}
	return &link
}
