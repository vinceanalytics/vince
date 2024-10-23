package web

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/web/db"
)

func LoginForm(db *db.Config, w http.ResponseWriter, r *http.Request) {
	db.HTML(w, login, nil)
}

func Login(db *db.Config, w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	password := r.Form.Get("password")
	if !db.PasswordMatch(password) {
		db.SaveCsrf(w)
		valid := map[string]any{
			"error": "Wrong password. Please try again.",
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
