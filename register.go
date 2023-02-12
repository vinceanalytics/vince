package vince

import (
	"context"
	"net/http"
	"strings"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/auth"
	"github.com/gernest/vince/email"
	"github.com/gernest/vince/models"
)

func (v *Vince) registerForm() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ServeHTML(w, templates.Register, http.StatusOK, templates.New(r.Context()))
	})
}

func (v *Vince) register() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, r := v.clientSession.Load(r)
		r.ParseForm()
		u, m, err := auth.NewUser(r)
		if err != nil {
			xlg.Err(err).Msg("Failed decoding new user from")
			ServeError(w, http.StatusInternalServerError)
			return
		}
		validCaptcha := session.VerifyCaptchaSolution(r.Form.Get(captchaKey))
		if len(m) > 0 || !validCaptcha {
			r = saveCsrf(w, r, session)
			r = saveCaptcha(w, r, session)
			if !validCaptcha {
				if m == nil {
					m = make(map[string]string)
				}
				m["captcha"] = "Please complete the captcha to register"
			}
			ServeHTML(w, templates.Register, http.StatusOK, templates.New(
				r.Context(),
				func(c *templates.Context) {
					c.Errors = m
					c.Form = r.Form
				},
			))
			return
		}
		if err := v.sql.Save(u).Error; err != nil {
			xlg.Err(err).Msg("failed saving new user")
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				r = saveCsrf(w, r, session)
				r = saveCaptcha(w, r, session)
				ServeHTML(w, templates.Register, http.StatusOK, templates.New(
					r.Context(),
					func(c *templates.Context) {
						c.Errors = map[string]string{
							"email": "already exists",
						}
						c.Form = r.Form
					},
				))
				return
			}
			ServeError(w, http.StatusInternalServerError)
			return
		}
		ctx := models.SetCurrentUser(r.Context(), u)
		session.Data[models.CurrentUserID] = u.ID
		session.Data["logged_in"] = true
		session.Save(w)
		if u.EmailVerified {
			http.Redirect(w, r, "/new", http.StatusFound)
		} else {
			err := v.sendVerificationEmail(ctx, u)
			if err != nil {
				xlg.Err(err).Msg("failed sending email message")
			}
			http.Redirect(w, r, "/activate", http.StatusFound)
		}
	})
}

func (v *Vince) sendVerificationEmail(ctx context.Context, usr *models.User) error {
	code, err := auth.IssueEmailVerification(v.sql, usr)
	if err != nil {
		return err
	}
	ctx = auth.SetActivationCode(ctx, code)
	return email.SendActivation(ctx, v.mailer)
}
