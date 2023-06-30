package site

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/templates"
	"github.com/vinceanalytics/vince/internal/timeseries"
	"github.com/vinceanalytics/vince/pkg/property"
	"github.com/vinceanalytics/vince/pkg/timex"
)

var homeTpl = templates.App("site/home.html")

func Home(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	now := core.Now(ctx)
	u := models.GetUser(ctx)
	site := models.GetSite(ctx)
	q := r.URL.Query()
	p := timex.Parse(q.Get("w"))
	m := property.ParsMetric(q.Get("m"))
	o := templates.SiteStats{
		Site:   site,
		Owner:  u.Name,
		Metric: m,
		Window: p,
		Global: timeseries.Global(ctx, u.ID, site.ID),
		Series: timeseries.QueryGlobalMetric(
			ctx, m, site.UserID, site.ID, p.Window(now), 0,
		),
	}
	render.HTML(ctx, w, homeTpl, http.StatusOK, func(ctx *templates.Context) {
		ctx.USER = u
		ctx.Site = site
		ctx.Stats = &o
	})
}
