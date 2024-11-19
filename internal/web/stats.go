package web

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/util/oracle"
	"github.com/vinceanalytics/vince/internal/web/db"
)

func Stats(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	if site.Locked {
		db.HTML(w, statsLocked, nil)
		return
	}
	hasStats := db.Ops().SeenFirstStats(site.Domain)
	w.Header().Set("x-robots-tag", "noindex, nofollow")
	db.HTML(w, stats, map[string]any{
		"seen_first_stats": hasStats,
		"title":            "vince · " + site.Domain,
		"demo":             site.Domain == oracle.Demo,
	})
}
