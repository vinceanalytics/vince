package auth

import (
	"net/http"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/render"
)

func UserSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	render.HTML(ctx, w, templates.UserSettings, http.StatusOK, func(ctx *templates.Context) {})
}
