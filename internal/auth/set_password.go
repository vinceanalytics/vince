package auth

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/render"
	"github.com/vinceanalytics/vince/sessions"
	"github.com/vinceanalytics/vince/templates"
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
