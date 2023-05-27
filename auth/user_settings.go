package auth

import (
	"encoding/base64"
	"net/http"

	"github.com/gernest/vince/models"
	"github.com/gernest/vince/pkg/secrets"
	"github.com/gernest/vince/render"
	"github.com/gernest/vince/templates"
)

func UserSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	usr := models.GetUser(ctx)
	models.PreloadUser(ctx, usr, "APIKeys")
	render.HTML(ctx, w, templates.UserSettings, http.StatusOK, func(ctx *templates.Context) {
		ctx.Key = base64.StdEncoding.EncodeToString(secrets.APIKey())
		ctx.CurrentUser = usr
	})
}
