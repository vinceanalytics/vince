package auth

import (
	"net/http"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/pkg/secrets"
	"github.com/gernest/vince/render"
)

func UserSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	usr := models.GetUser(ctx)
	models.PreloadUser(ctx, usr, "APIKeys")
	render.HTML(ctx, w, templates.UserSettings, http.StatusOK, func(ctx *templates.Context) {
		ctx.Key = secrets.GenerateAPIKey()
		ctx.CurrentUser = usr
	})
}
