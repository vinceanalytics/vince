package stats

import (
	"encoding/json"
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/pkg/log"
	"github.com/vinceanalytics/vince/timeseries"
)

func Query(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := models.GetSite(ctx)
	u := models.SiteOwner(ctx, site.ID)
	var base timeseries.BaseQuery
	err := json.NewDecoder(r.Body).Decode(&base)
	if err != nil {
		log.Get().Err(err).Msg("failed to decode query body")
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if err := base.Filters.Validate(); err != nil {
		log.Get().Err(err).Msg("failed query filters validation")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	o := timeseries.Query(ctx, timeseries.QueryRequest{
		UserID:    u.ID,
		SiteID:    site.ID,
		BaseQuery: base,
	})
	render.JSON(w, http.StatusOK, o)
}
