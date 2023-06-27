package site

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/templates"
)

var settingsTpl = templates.App("site/settings.html")

func Settings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := models.GetUser(ctx)
	site := models.GetSite(ctx)
	models.PreloadSite(ctx, site, "Goals")
	render.HTML(ctx, w, settingsTpl, http.StatusOK, func(ctx *templates.Context) {
		ctx.Site = site
		ctx.USER = u
	})
}
