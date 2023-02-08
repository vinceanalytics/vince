package vince

import (
	"net/http"

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
		r.ParseForm()
		u := new(models.User)
		m := v.DecodeRegistrationForm(u, r)
		ctx := r.Context()
		if len(m) > 0 {
			ServeHTML(w, templates.Register, http.StatusOK, templates.New(
				ctx,
				func(c *templates.Context) {
					c.Errors = m
					c.Form = r.Form
				},
			))
			return
		}
	})
}
