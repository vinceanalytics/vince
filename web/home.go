package web

import (
	"net/http"

	"github.com/gernest/len64/web/db"
)

func Home(db *db.Config, w http.ResponseWriter, r *http.Request) {
	home.Execute(w, map[string]any{})
}
