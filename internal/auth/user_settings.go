package auth

import (
	"encoding/base64"
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/templates"
	"github.com/vinceanalytics/vince/pkg/secrets"
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