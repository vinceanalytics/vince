package site

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/render"
	"github.com/vinceanalytics/vince/templates"
)

func New(w http.ResponseWriter, r *http.Request) {
	u := models.GetUser(r.Context())
	owned := models.CountOwnedSites(r.Context(), u.ID)
	render.HTML(r.Context(), w, templates.SiteNew, http.StatusOK, func(ctx *templates.Context) {
		ctx.NewSite = &templates.NewSite{
			IsFirstSite: owned == 0,
		}
	})
}
