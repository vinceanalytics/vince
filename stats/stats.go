package stats

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/gernest/vince/models"
	"github.com/gernest/vince/pkg/log"
	"github.com/gernest/vince/render"
	"github.com/gernest/vince/timeseries"
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
	if base.Match.IsRe {
		base.Match.Re, err = regexp.Compile(base.Match.Text)
		if err != nil {
			log.Get().Err(err).Msg("failed to compile query match re")
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
	}
	o := timeseries.Query(ctx, timeseries.QueryRequest{
		UserID:    u.ID,
		SiteID:    site.ID,
		BaseQuery: base,
	})
	render.JSON(w, http.StatusOK, o)
}
