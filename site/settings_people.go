package site

import (
	"net/http"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/render"
)

func SettingsPeople(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := models.GetSite(ctx)
	models.PreloadSite(ctx, site, "SiteMemberships", "SiteMemberships.User", "Invitations")
	render.HTML(ctx, w, templates.SiteSettingsPeople, http.StatusOK, func(ctx *templates.Context) {
		ctx.Site = site
		ctx.Page = " people"
	})
}
