package conversions

import (
	"math"
	"net/http"

	"github.com/vinceanalytics/vince/internal/api/aggregates"
	"github.com/vinceanalytics/vince/internal/api/breakdown"
	"github.com/vinceanalytics/vince/internal/web/db"
	"github.com/vinceanalytics/vince/internal/web/query"
)

func Conversion(db *db.Config, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	sx := aggregates.
		Aggregates(
			ctx, db.TimeSeries(),
			site.Domain, params.Start(), params.End(), params.Interval(), params.Filter(), []string{"visitors"})

	sx.Compute()
	totalVisitors := sx.Visitors
	rs, err := breakdown.BreakdownGoals(ctx, db.TimeSeries(), site, params, []string{"visitors", "events"})
	if err != nil {
		db.Logger().Error("breakdown goals", "err", err)
		rs = new(breakdown.Result)
	}
	for i := range rs.Results {
		m := rs.Results[i]
		m["conversion_rate"] = coversionRate(totalVisitors, m["visitors"].(float64))
	}
	db.JSON(w, rs)
}

func coversionRate(uniq, converted float64) float64 {
	if uniq > 0 {
		return math.Round(converted / uniq * 100)
	}
	return 0
}
