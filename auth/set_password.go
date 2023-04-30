package auth

import (
	"net/http"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/render"
	"github.com/gernest/vince/sessions"
)

func SetPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fail := models.SetPassword(ctx, r.Form.Get("password"))
	if fail != "" {
		r = sessions.SaveCsrf(w, r)
		render.HTML(ctx, w, templates.PasswordForm, http.StatusOK, func(ctx *templates.Context) {
			ctx.Errors["password"] = fail
			ctx.Form = r.Form
		})
		return
	}
	http.Redirect(w, r, "/sites/new", http.StatusFound)
}
