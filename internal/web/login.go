package web

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/config"
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
	if config.C.Admin.Email != email || !ro2.PasswordMatch(password) {
		db.SaveCsrf(w)
		db.SaveCaptcha(w)
		valid := map[string]any{
			"error": "Wrong email or password. Please try again.",
		}
		db.HTML(w, login, valid)
		return
	}
	http.Redirect(w, r, db.Login(w), http.StatusFound)
}

func Logout(db *db.Config, w http.ResponseWriter, r *http.Request) {
	db.Logout(w)
	http.Redirect(w, r, "/", http.StatusFound)
}
