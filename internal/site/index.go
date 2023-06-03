package site

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/templates"
)

func Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := models.GetUser(ctx)
	models.PreloadUser(ctx, u, "Sites")
	render.HTML(ctx, w, templates.Sites, http.StatusOK, func(ctx *templates.Context) {
		ctx.SitesOverview = make([]models.SiteOverView, len(u.Sites))
		for i := range u.Sites {
			ctx.SitesOverview[i] = models.SiteOverView{
				Site: u.Sites[i],
			}
		}
	})
}
