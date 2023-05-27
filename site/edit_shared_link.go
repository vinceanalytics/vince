package site

import (
	"net/http"

	"github.com/gernest/vince/models"
	"github.com/gernest/vince/params"
	"github.com/gernest/vince/render"
	"github.com/gernest/vince/templates"
)

func EditSharedLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := models.GetSite(ctx)
	shared := models.GetSharedLinkWithSlug(ctx, site.ID, params.Get(ctx)["slug"])
	render.HTML(ctx, w, templates.EditSharedLinkForm, http.StatusOK, func(ctx *templates.Context) {
		ctx.Site = site
		ctx.SharedLink = shared
	})
}
