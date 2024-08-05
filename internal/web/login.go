package web

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/kv"
	"github.com/vinceanalytics/vince/internal/web/db"
)

func LoginForm(db *db.Config, w http.ResponseWriter, r *http.Request) {
	db.HTML(w, login, nil)
}

func Login(db *db.Config, w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	u := new(kv.User)
	err := u.ByEmail(db.Get(), email)
	if err != nil {
		db.SaveCsrf(w)
		db.SaveCaptcha(w)
		valid := map[string]any{
			"error": "Wrong email or password. Please try again.",
		}
		db.HTML(w, login, valid)
		db.Logger().Error("login", "err", err)
		return
	}

	if !u.PasswordMatch(password) {
		db.SaveCsrf(w)
		db.SaveCaptcha(w)
		valid := map[string]any{
			"error": "Wrong email or password. Please try again.",
		}
		db.HTML(w, login, valid)
		return
	}

	http.Redirect(w, r, db.Login(w, u.ID()), http.StatusFound)
}
