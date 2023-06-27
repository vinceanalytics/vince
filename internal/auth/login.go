package auth

import (
	"fmt"
	"net/http"

	"github.com/vinceanalytics/vince/internal/caches"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/remoteip"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/sessions"
	"github.com/vinceanalytics/vince/internal/templates"
)

var loginForm = templates.Focus("auth/login_form.html")

func Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if !caches.AllowRemoteIPLogin(ctx, remoteip.Get(r)) {
		render.ERROR(ctx, w, http.StatusTooManyRequests, func(ctx *templates.Context) {
			ctx.Error.StatusText = "Too many login attempts. Wait a minute before trying again."
		})
		return
	}
	nameOrEmail := r.Form.Get("name-or-email")
	password := r.Form.Get("password")
	u := models.QueryUserByNameOrEmail(ctx, nameOrEmail)
	session, r := sessions.Load(r)
	validCaptcha := session.VerifyCaptchaSolution(r)

	if !validCaptcha || u == nil {
		r = sessions.SaveCsrf(w, r)
		r = sessions.SaveCaptcha(w, r)
		render.HTML(r.Context(), w, loginForm, http.StatusOK, func(ctx *templates.Context) {
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
	if !models.PasswordMatch(u, password) {
		r = sessions.SaveCsrf(w, r)
		r = sessions.SaveCaptcha(w, r)
		render.HTML(r.Context(), w, loginForm, http.StatusOK, func(ctx *templates.Context) {
			ctx.Errors["failed"] = "Wrong email or password. Please try again."
			ctx.Form = r.Form
		})
		return
	}
	if session.Data.LoginDest == "" {
		session.Data.LoginDest = fmt.Sprintf("/%s", u.Name)
	}
	session.Data.USER = u.Name
	session.Data.LoggedIn = true
	dest := session.Data.LoginDest
	session.Data.LoginDest = ""
	session.Save(ctx, w)
	http.Redirect(w, r, dest, http.StatusFound)
}
