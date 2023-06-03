package site

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/caches"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/sessions"
	"github.com/vinceanalytics/vince/internal/timeseries"
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
	timeseries.GetMap(ctx).Delete(site.ID)

	// permanently remove site stats
	timeseries.DropSite(ctx, u.ID, site.ID)

	session, r := sessions.Load(r)
	session.SuccessFlash("site and site stats have been permanently deleted").Save(ctx, w)
	http.Redirect(w, r, "/sites", http.StatusFound)
}
