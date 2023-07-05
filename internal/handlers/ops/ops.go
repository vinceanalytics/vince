package ops

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/templates"
)

var systemTpl = templates.App("user/system.html")

func System(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	usr := models.GetUser(ctx)
	render.HTML(ctx, w, systemTpl, http.StatusOK, func(ctx *templates.Context) {
		ctx.USER = usr
		ctx.Header.Context = "system"
		ctx.Header.ContextRef = "/system"
	})
}
