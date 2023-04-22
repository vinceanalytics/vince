package auth

import (
	"net/http"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/caches"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/remoteip"
	"github.com/gernest/vince/render"
	"github.com/gernest/vince/sessions"
)

func Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if !caches.AllowRemoteIPLogin(ctx, remoteip.Get(r)) {
		render.ERROR(ctx, w, http.StatusTooManyRequests, func(ctx *templates.Context) {
			ctx.Error.StatusText = "Too many login attempts. Wait a minute before trying again."
		})
		return
	}
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	u := models.UserByEmail(ctx, email)
	session, r := sessions.Load(r)

	validCaptcha := session.VerifyCaptchaSolution(r)

	if !validCaptcha || u == nil {
		r = sessions.SaveCsrf(w, r)
		r = sessions.SaveCaptcha(w, r)
		render.HTML(r.Context(), w, templates.LoginForm, http.StatusOK, func(ctx *templates.Context) {
			if !validCaptcha {
				ctx.Errors["captcha"] = "Invalid Captcha"
			}
			if u == nil {
				ctx.Errors["failed"] = "Wrong email or password. Please try again."
			}
			ctx.Form = r.Form
		})
		return
	}
	if !caches.AllowUseIDToLogin(ctx, u.ID) {
		render.ERROR(ctx, w, http.StatusTooManyRequests, func(ctx *templates.Context) {
			ctx.Error.StatusText = "Too many login attempts. Wait a minute before trying again."
		})
		return
	}
	if !u.PasswordMatch(password) {
		r = sessions.SaveCsrf(w, r)
		r = sessions.SaveCaptcha(w, r)
		render.HTML(r.Context(), w, templates.LoginForm, http.StatusOK, func(ctx *templates.Context) {
			ctx.Errors["failed"] = "Wrong email or password. Please try again."
			ctx.Form = r.Form
		})
		return
	}
	if session.Data.LoginDest == "" {
		session.Data.LoginDest = "/sites"
	}
	session.Data.CurrentUserID = u.ID
	session.Data.LoggedIn = true
	dest := session.Data.LoginDest
	session.Data.LoginDest = ""
	session.Save(w)
	http.Redirect(w, r, dest, http.StatusFound)
}
