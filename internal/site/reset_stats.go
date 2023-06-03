package site

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/flash"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/sessions"
	"github.com/vinceanalytics/vince/timeseries"
)

func ResetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := models.GetSite(ctx)
	u := models.GetUser(ctx)
	// permanently remove site stats
	timeseries.DropSite(ctx, u.ID, site.ID)
	session, r := sessions.Load(r)
	session.Data.Flash = &flash.Flash{
		Success: []string{"site stats have been permanently deleted"},
	}
	session.Save(ctx, w)
	http.Redirect(w, r, "/sites", http.StatusFound)

}
