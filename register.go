package vince

import (
	"net/http"
	"strings"

	"github.com/gernest/vince/assets/ui/templates"
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
		u := new(models.User)
		m := v.DecodeRegistrationForm(u, r)
		validCaptcha := session.VerifyCaptchaSolution(r.Form.Get(captchaKey))
		if len(m) > 0 || !validCaptcha {
			r = saveCsrf(session, w, r)
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
				r = saveCsrf(session, w, r)
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
		session.Data[models.CurrentUserID] = u.ID
		session.Data["logged_in"] = true
		session.Save(w)
		http.Redirect(w, r, "/login", http.StatusFound)
	})
}
