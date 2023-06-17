package user

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/templates"
)

func Profile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := models.GetUser(ctx)
	render.HTML(ctx, w, templates.Profile, http.StatusOK, func(ctx *templates.Context) {
		ctx.CurrentUser = u
		ctx.Header.Context = u.Name
		ctx.Header.ContextRef = "/" + u.Name

	})
}
