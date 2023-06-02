package site

import (
	"net/http"

	"github.com/vinceanalytics/vince/models"
	"github.com/vinceanalytics/vince/render"
	"github.com/vinceanalytics/vince/templates"
)

func NewGoal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := models.GetSite(ctx)
	render.HTML(ctx, w, templates.SiteNewGoal, http.StatusOK, func(ctx *templates.Context) {
		ctx.Site = site
	})
}
