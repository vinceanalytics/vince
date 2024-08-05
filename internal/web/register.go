package web

import (
	"net/http"
	"strings"

	"github.com/vinceanalytics/vince/internal/kv"
	"github.com/vinceanalytics/vince/internal/web/db"
)

func RegisterForm(db *db.Config, w http.ResponseWriter, r *http.Request) {
	db.HTML(w, register, nil)
}

func Register(db *db.Config, w http.ResponseWriter, r *http.Request) {
	usr := new(kv.User)
	m, err := usr.NewUser(r)
	if err != nil {
		db.HTMLCode(http.StatusInternalServerError, w, e500, nil)
		db.Logger().Error("creating new user", "err", err)
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
		db.HTML(w, register, m)
		return
	}
	err = usr.Save(db.Get())
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			db.SaveCsrf(w)
			db.SaveCaptcha(w)
			db.HTML(w, register, map[string]any{
				"validation.email": "email already exists",
			})
			return
		}
		db.HTMLCode(http.StatusInternalServerError, w, e500, nil)
		return
	}
	http.Redirect(w, r, "/sites/new", http.StatusFound)
}
