package vince

import (
	"net/http"

	"github.com/gernest/vince/assets/ui/templates"
)

func (v *Vince) registerForm() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ServeHTML(w, templates.Register, http.StatusOK, map[string]any{
			"csrf":    getCsrf(r.Context()),
			"captcha": getCaptcha(r.Context()),
		})
	})
}

func (v *Vince) register() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		u := new(User)
		m := v.DecodeRegistrationForm(u, r)
		ctx := r.Context()
		if len(m) > 0 {
			// render the registration form with errors
			ServeHTML(w, templates.Register, http.StatusOK, map[string]any{
				"csrf":    getCsrf(ctx),
				"captcha": getCaptcha(ctx),
				"errors":  m,
				"form":    r.Form,
			})
			return
		}
	})
}
