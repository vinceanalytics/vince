package web

import (
	"net/http"
	"strings"

	"github.com/gernest/len64/web/db"
	"github.com/gernest/len64/web/db/schema"
)

func RegisterForm(db *db.Config, w http.ResponseWriter, r *http.Request) {
	register.Execute(w, db.Context(make(map[string]any)))
}

func Register(db *db.Config, w http.ResponseWriter, r *http.Request) {
	usr := new(schema.User)
	m, err := usr.NewUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	validCaptcha := db.VerifyCaptchaSolution(r)
	if len(m) > 0 || !validCaptcha {
		db.SaveCsrf(w)
		db.SaveCaptcha(w)
		if len(m) == 0 {
			m = make(map[string]any)
		}
		if !validCaptcha {
			m["validation.captcha"] = "invalid captcha"
		}
		register.Execute(w, db.Context(m))
		return
	}
	err = usr.Save(db.Get())
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			db.SaveCsrf(w)
			db.SaveCaptcha(w)
			register.Execute(w, db.Context(map[string]any{
				"validation.email": "email already exists",
			}))
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/new", http.StatusFound)
}
