package site

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/templates"
)

func Home(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := models.GetUser(ctx)
	site := models.GetSite(ctx)
	render.HTML(ctx, w, templates.SiteHome, http.StatusOK, func(ctx *templates.Context) {
		ctx.USER = u
		ctx.Site = site
	})
}
