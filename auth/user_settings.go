package auth

import (
	"net/http"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/render"
)

func UserSettings(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/settings/profile", http.StatusFound)
}

func UserSettingsProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	render.HTML(ctx, w, templates.UserSettingsProfile, http.StatusOK, func(ctx *templates.Context) {
		ctx.Page = " profile"
	})
}

func UserSettingsAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	render.HTML(ctx, w, templates.UserSettingsAccount, http.StatusOK, func(ctx *templates.Context) {
		ctx.Page = " account"
	})
}

func UserSettingsAPIKeys(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	render.HTML(ctx, w, templates.UserSettingsAPIKeys, http.StatusOK, func(ctx *templates.Context) {
		ctx.Page = " api_keys"
	})
}
