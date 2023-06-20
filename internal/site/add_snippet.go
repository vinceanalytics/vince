package site

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/templates"
)

func AddSnippet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := models.GetSite(ctx)
	render.HTML(ctx, w, templates.AddSnippet, http.StatusOK, func(ctx *templates.Context) {
		ctx.Site = site
		ctx.IsFIrstSite = false
	})
}
