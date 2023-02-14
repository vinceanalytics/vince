package stats

import (
	"net/http"
	"net/url"

	"github.com/gernest/vince/log"
	"github.com/gernest/vince/render"
	"github.com/gernest/vince/timeseries"
)

func V1RealtimeVisitors(w http.ResponseWriter, r *http.Request) {
	query := make(url.Values)
	query.Set("period", "realtime")
	if res, err := timeseries.Get(r.Context()).CurrentVisitors(r.Context(), timeseries.QueryFrom(query)); err != nil {
		log.Get(r.Context()).Err(err).Msg("failed to query events")
		render.JSON(w, http.StatusInternalServerError, timeseries.Record{})
	} else {
		render.JSON(w, http.StatusOK, res)
	}
}

func V1Aggregate(w http.ResponseWriter, r *http.Request) {
}
func V1Breakdown(w http.ResponseWriter, r *http.Request) {
}
func V1Timeseries(w http.ResponseWriter, r *http.Request) {
}
