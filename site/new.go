package site

import (
	"net/http"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/render"
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
