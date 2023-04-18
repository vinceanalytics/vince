package sites

import (
	"net/http"

	"github.com/gernest/vince/caches"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/render"
)

func Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := models.GetSite(ctx)

	// remove site from database
	err := models.Get(ctx).Delete(site).Error
	if err != nil {
		models.DBE(ctx, err, "failed to delete site")
	}
	// remove from cache
	caches.Site(ctx).Del(site.Domain)
	render.ERROR(r.Context(), w, http.StatusNotImplemented)
}
