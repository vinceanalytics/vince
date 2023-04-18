package site

import (
	"net/http"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/render"
)

func SettingsCustomDomain(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := models.GetSite(ctx)
	render.HTML(ctx, w, templates.SiteSettingsCustomDomain, http.StatusOK, func(ctx *templates.Context) {
		ctx.Site = site
		ctx.Page = " custom-domain"
	})
}
