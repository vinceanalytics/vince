package site

import (
	"net/http"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/plot"
	"github.com/gernest/vince/render"
)

func Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := models.GetUser(ctx)
	u.Preload(ctx, "Sites")
	render.HTML(ctx, w, templates.Sites, http.StatusOK, func(ctx *templates.Context) {
		ctx.Page = "sites"
		ctx.SitesOverview = make([]models.SiteOverView, len(u.Sites))
		for i := range u.Sites {
			ctx.SitesOverview[i] = models.SiteOverView{
				Site: u.Sites[i],
				Plot: &plot.U{
					ID:     u.Sites[i].ID,
					Width:  1613,
					Height: 240,
					Series: []float64{0, 13, 11, 4, 44, 10},
				},
			}
		}
	})
}
