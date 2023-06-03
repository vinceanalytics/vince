package site

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/templates"
	"gorm.io/gorm"
)

func AddSnippet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := models.GetUser(ctx)
	site := models.GetSite(ctx)
	isFirst := !models.Exists(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Model(&models.SiteMembership{}).
			Where("site_id", site.ID).
			Where("user_id", user.ID)
	})
	render.HTML(ctx, w, templates.AddSnippet, http.StatusOK, func(ctx *templates.Context) {
		ctx.Site = site
		ctx.IsFIrstSite = isFirst
	})
}
