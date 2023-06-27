package goals

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/templates"
)

func New(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := models.GetUser(ctx)
	site := models.GetSite(ctx)
	render.HTML(ctx, w, newGoalTpl, http.StatusOK, func(ctx *templates.Context) {
		ctx.Site = site
		ctx.USER = u
	})
}
