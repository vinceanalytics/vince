package web

import (
	"net/http"

	"github.com/gernest/len64/web/db"
)

func RegisterForm(db *db.Config, w http.ResponseWriter, r *http.Request) {
	register.Execute(w, db.Context(make(map[string]any)))
}
