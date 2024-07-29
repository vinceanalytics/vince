package web

import (
	"net/http"

	"github.com/gernest/len64/web/db"
)

func CreateSiteForm(db *db.Config, w http.ResponseWriter, r *http.Request) {
	createSite.Execute(w, db.Context(make(map[string]any)))
}

func CreateSite(db *db.Config, w http.ResponseWriter, r *http.Request) {
	createSite.Execute(w, db.Context(make(map[string]any)))
}
