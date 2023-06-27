package account

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/templates"
)

var settingsTpl = templates.App("auth/user_settings.html")

func Settings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	usr := models.GetUser(ctx)
	models.PreloadUser(ctx, usr, "APIKeys")
	render.HTML(ctx, w, settingsTpl, http.StatusOK, func(ctx *templates.Context) {
		ctx.Header = templates.Header{
			Context:    "Settings",
			ContextRef: "/settings#profile",
		}
		ctx.USER = usr
	})
}
