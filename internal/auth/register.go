package auth

import (
	"net/http"
	"strings"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/sessions"
	"github.com/vinceanalytics/vince/internal/templates"
	"github.com/vinceanalytics/vince/pkg/log"
)

func Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session, r := sessions.Load(r)
	r.ParseForm()
	u := &models.User{
		Name:     r.Form.Get("name"),
		FullName: r.Form.Get("full-name"),
		Email:    r.Form.Get("email"),
	}
	m := models.NewUser(ctx, u, r.Form.Get("password"), r.Form.Get("password_confirmation"))
	validCaptcha := session.VerifyCaptchaSolution(r)
	if len(m) > 0 || !validCaptcha {
		r = sessions.SaveCsrf(w, r)
		r = sessions.SaveCaptcha(w, r)
		render.HTML(r.Context(), w, templates.RegisterForm, http.StatusOK, func(ctx *templates.Context) {
			for k, v := range m {
				ctx.Errors[k] = v
			}
			if !validCaptcha {
				ctx.Errors["captcha"] = "Please complete the captcha to register"
			}
			ctx.Form = r.Form
		})
		return
	}
	if err := models.Get(r.Context()).Save(u).Error; err != nil {
		log.Get().Err(err).Msg("failed saving new user")
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			r = sessions.SaveCsrf(w, r)
			r = sessions.SaveCaptcha(w, r)
			render.HTML(r.Context(), w, templates.RegisterForm, http.StatusOK, func(ctx *templates.Context) {
				ctx.Errors["email"] = "already exists"
				ctx.Form = r.Form
			})
			return
		}
		render.ERROR(r.Context(), w, http.StatusInternalServerError)
		return
	}
	ctx = models.SetUser(ctx, u)
	session.Data.USER = u.Name
	session.Data.LoggedIn = true
	session.Save(ctx, w)
	if u.EmailVerified {
		http.Redirect(w, r, "/new", http.StatusFound)
	} else {
		err := SendVerificationEmail(ctx, u)
		if err != nil {
			log.Get().Err(err).Msg("failed sending email message")
		}
		http.Redirect(w, r, "/activate", http.StatusFound)
	}
}
