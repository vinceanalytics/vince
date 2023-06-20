package sites

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/caches"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/params"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/timeseries"
)

func DeleteSite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := models.GetUser(ctx)
	site := models.SiteFor(ctx,
		models.GetUser(ctx).ID,
		params.Get(ctx)["site_id"],
	)
	if site != nil {
		// remove site from database
		models.DeleteSite(ctx, site)

		// remove site from cache
		caches.Site(ctx).Del(site.Domain)

		// remove site events in collection  buffers
		timeseries.GetMap(ctx).Delete(site.ID)

		// permanently remove site stats
		timeseries.DropSite(ctx, u.ID, site.ID)
		render.JSON(w, http.StatusOK, map[string]any{
			"deleted": true,
		})
		return
	}
	render.JSON(w, http.StatusNotFound, map[string]any{
		"error": http.StatusText(http.StatusNotFound),
	})
}
