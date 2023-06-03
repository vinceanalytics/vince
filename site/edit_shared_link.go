package site

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/params"
	"github.com/vinceanalytics/vince/render"
	"github.com/vinceanalytics/vince/templates"
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
