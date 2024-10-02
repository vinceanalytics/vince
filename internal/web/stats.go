package web

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/web/db"
)

func Stats(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	if site.Locked {
		db.HTML(w, statsLocked, nil)
		return
	}
	db.HTML(w, stats, map[string]any{
		"title": "vince · " + site.Domain,
	})
}
