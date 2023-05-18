package site

import (
	"net/http"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/render"
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
