package web

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/ro2"
	"github.com/vinceanalytics/vince/internal/web/db"
)

func LoginForm(db *db.Config, w http.ResponseWriter, r *http.Request) {
	db.HTML(w, login, nil)
}

func Login(db *db.Config, w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	u := db.Get().UserByEmail(email)
	if u == nil {
		db.SaveCsrf(w)
		db.SaveCaptcha(w)
		valid := map[string]any{
			"error": "Wrong email or password. Please try again.",
		}
		db.HTML(w, login, valid)
		return
	}

	if !ro2.PasswordMatch(u, password) {
		db.SaveCsrf(w)
		db.SaveCaptcha(w)
		valid := map[string]any{
			"error": "Wrong email or password. Please try again.",
		}
		db.HTML(w, login, valid)
		return
	}

	http.Redirect(w, r, db.Login(w, ro2.ID(u.Id)), http.StatusFound)
}
