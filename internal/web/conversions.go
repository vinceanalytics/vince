package web

import (
	"math"
	"net/http"

	"github.com/vinceanalytics/vince/internal/ro2"
	"github.com/vinceanalytics/vince/internal/web/db"
	"github.com/vinceanalytics/vince/internal/web/query"
)

func Conversion(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	sx, err := db.Get().Aggregates(site.Domain, params.Start(), params.End(), params.Interval(), params.Filter(), []string{"visitors"})
	if err != nil {
		db.Logger().Error("reading aggregates", "err", err)
		sx = new(ro2.Stats)
	}
	sx.Compute()
	totalVisitors := sx.Visitors
	rs, err := db.Get().BreakdownGoals(site, params, []string{"visitors", "events"})
	if err != nil {
		db.Logger().Error("breakdown goals", "err", err)
		rs = new(ro2.Result)
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
