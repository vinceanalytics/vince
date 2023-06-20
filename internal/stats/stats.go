package stats

import (
	"encoding/json"
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/query"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/timeseries"
	"github.com/vinceanalytics/vince/pkg/log"
)

func Query(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := models.GetSite(ctx)
	var q query.Query
	err := json.NewDecoder(r.Body).Decode(&q)
	if err != nil {
		log.Get().Err(err).Msg("failed to decode query body")
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	render.JSON(w, http.StatusOK, timeseries.Query(ctx, site.UserID, site.ID, q))
}
