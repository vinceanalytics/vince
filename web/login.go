package web

import (
	"net/http"

	"github.com/gernest/len64/web/db"
)

func LoginForm(db *db.Config, w http.ResponseWriter, r *http.Request) {
	login.Execute(w, map[string]any{})
}
