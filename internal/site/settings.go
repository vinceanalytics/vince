package site

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/templates"
)

func Settings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := models.GetSite(ctx)
	models.PreloadSite(ctx, site, "SiteMemberships", "SiteMemberships.User", "Invitations", "SharedLinks")
	goals := models.Goals(ctx, site.Domain)
	render.HTML(ctx, w, templates.SiteSettings, http.StatusOK, func(ctx *templates.Context) {
		ctx.Site = site
		ctx.Goals = goals
	})
}
