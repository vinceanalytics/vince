package site

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/templates"
	"github.com/vinceanalytics/vince/internal/timeseries"
)

func Home(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := models.GetUser(ctx)
	site := models.GetSite(ctx)
	o := templates.SiteStats{
		Site:   site,
		Owner:  u.Name,
		Global: timeseries.Global(ctx, u.ID, site.ID),
	}
	render.HTML(ctx, w, templates.SiteHome, http.StatusOK, func(ctx *templates.Context) {
		ctx.USER = u
		ctx.Site = site
		ctx.Stats = &o
	})
}
