package vince

import (
	"net/http"

	"github.com/gernest/vince/stats"
)

func v1Stats() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			switch r.URL.Path {
			case "/api/v1/stats/realtime/visitors":
				stats.V1RealtimeVisitors(w, r)
				return
			case "/api/v1/stats/aggregate":
				stats.V1Aggregate(w, r)
				return
			case "/api/v1/stats/breakdown":
				stats.V1Breakdown(w, r)
				return
			case "/api/v1/stats/timeseries":
				stats.V1Timeseries(w, r)
				return
			}
		}
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})
}
