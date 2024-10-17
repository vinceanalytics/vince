package web

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/web/db"
)

func Conversion(db *db.Config, w http.ResponseWriter, r *http.Request) {
	// site := db.CurrentSite()
	// params := query.New(r.URL.Query())
	// sx, err := db.Get().Aggregates(site.Domain, params.Start(), params.End(), params.Interval(), params.Filter(), []string{"visitors"})
	// if err != nil {
	// 	db.Logger().Error("reading aggregates", "err", err)
	// 	sx = new(ro2.Stats)
	// }
	// sx.Compute()
	db.JSON(w, map[string]any{
		"results": []any{},
	})
}
