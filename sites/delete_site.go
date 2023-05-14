package sites

import (
	"net/http"

	"github.com/gernest/vince/caches"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/params"
	"github.com/gernest/vince/render"
	"github.com/gernest/vince/timeseries"
)

func DeleteSite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := models.GetUser(ctx)
	site := models.SiteFor(ctx,
		models.GetUser(ctx).ID,
		params.Get(ctx)["site_id"],
		"owner",
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
