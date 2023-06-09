package site

import (
	"context"
	"net/http"
	"time"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/templates"
	"github.com/vinceanalytics/vince/internal/timeseries"
)

func Index(w http.ResponseWriter, r *http.Request) {
	rctx := r.Context()
	u := models.GetUser(rctx)
	models.PreloadUser(rctx, u, "Sites")
	render.HTML(rctx, w, templates.Sites, http.StatusOK, func(ctx *templates.Context) {
		ctx.SitesOverview = make([]models.SiteOverView, len(u.Sites))
		for i := range u.Sites {
			ctx.SitesOverview[i] = models.SiteOverView{
				Site:     u.Sites[i],
				Visitors: visitors(rctx, u.ID, u.Sites[i].ID),
			}
		}
	})
}

func visitors(ctx context.Context, uid, sid uint64) string {
	q := timeseries.Root(ctx, uid, sid, timeseries.RootOptions{
		Metric:  timeseries.Visitors,
		Prop:    timeseries.Base,
		Window:  time.Hour * 24,
		NoProps: true,
	})
	a := q.Aggregate[timeseries.Base.String()][timeseries.Visitors.String()]
	for _, v := range a {
		if v.Key == timeseries.BaseKey {
			return v.Value.String()
		}
	}
	return "0"
}
