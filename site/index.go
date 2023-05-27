package site

import (
	"net/http"

	"github.com/gernest/vince/models"
	"github.com/gernest/vince/render"
	"github.com/gernest/vince/templates"
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
