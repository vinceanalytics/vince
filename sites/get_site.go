package sites

import (
	"net/http"

	"github.com/vinceanalytics/vince/models"
	"github.com/vinceanalytics/vince/params"
	"github.com/vinceanalytics/vince/render"
)

func GetSite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := models.SiteFor(ctx,
		models.GetUser(ctx).ID,
		params.Get(ctx)["site_id"],
		"owner", "admin",
	)
	if site != nil {
		render.JSON(w, http.StatusOK, site)
		return
	}
	render.JSON(w, http.StatusNotFound, map[string]any{
		"error": http.StatusText(http.StatusNotFound),
	})
}
