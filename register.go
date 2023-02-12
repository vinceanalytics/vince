package vince

import (
	"context"
	"net/http"
	"strings"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/auth"
	"github.com/gernest/vince/email"
	"github.com/gernest/vince/log"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/sessions"
)

func (v *Vince) registerForm() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ServeHTML(r.Context(), w, templates.Register, http.StatusOK, templates.New(r.Context()))
	})
}

func (v *Vince) register() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, r := sessions.Load(r)
		r.ParseForm()
		u, m, err := auth.NewUser(r)
		if err != nil {
			log.Get(r.Context()).Err(err).Msg("Failed decoding new user from")
			ServeError(r.Context(), w, http.StatusInternalServerError)
			return
		}

		validCaptcha := session.VerifyCaptchaSolution(r)
		if len(m) > 0 || !validCaptcha {
			r = sessions.SaveCsrf(w, r)
			r = sessions.SaveCaptcha(w, r)
			if !validCaptcha {
				if m == nil {
					m = make(map[string]string)
				}
				m["_captcha"] = "Please complete the captcha to register"
			}
			ServeHTML(r.Context(), w, templates.Register, http.StatusOK, templates.New(
				r.Context(),
				func(c *templates.Context) {
					c.Errors = m
					c.Form = r.Form
				},
			))
			return
		}
		if err := v.sql.Save(u).Error; err != nil {
			log.Get(r.Context()).Err(err).Msg("failed saving new user")
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				r = sessions.SaveCsrf(w, r)
				r = sessions.SaveCaptcha(w, r)
				ServeHTML(r.Context(), w, templates.Register, http.StatusOK, templates.New(
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
			ServeError(r.Context(), w, http.StatusInternalServerError)
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
				log.Get(r.Context()).Err(err).Msg("failed sending email message")
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
