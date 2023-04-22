package site

import (
	"net/http"

	"github.com/gernest/vince/caches"
	"github.com/gernest/vince/flash"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/sessions"
	"github.com/gernest/vince/timeseries"
)

func DeleteSite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := models.GetSite(ctx)
	u := models.GetUser(ctx)

	// remove site from database
	models.DeleteSite(ctx, site)

	// remove site from cache
	caches.Site(ctx).Del(site.Domain)

	// remove site events in collection  buffers
	timeseries.GetMap(ctx).Delete(ctx, site.ID)

	// permanently remove site stats
	timeseries.DropSite(ctx, u.ID, site.ID)
	session, r := sessions.Load(r)
	session.Data.Flash = &flash.Flash{
		Success: []string{"site and site stats have been permanently deleted"},
	}
	session.Save(w)
	http.Redirect(w, r, "/sites", http.StatusFound)
}
