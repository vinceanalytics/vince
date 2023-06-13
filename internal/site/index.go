package site

import (
	"context"
	"net/http"
	"time"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/query"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/templates"
	"github.com/vinceanalytics/vince/internal/timeseries"
	"github.com/vinceanalytics/vince/pkg/property"
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

func visitors(ctx context.Context, uid, sid uint64) uint32 {
	q := timeseries.Query(ctx, uid, sid, query.Query{
		Window: &query.Duration{
			Value: time.Hour * 24,
		},
		Props: &query.Props{
			Base: &query.Metrics{
				Visitors: &query.Select{
					Exact: &query.Value{
						Value: property.BaseKey,
					},
				},
			},
		},
	})
	return timeseries.Sum(
		q.Props.Base.Visitors[property.BaseKey],
	)
}
