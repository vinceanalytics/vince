package stats

import (
	"encoding/json"
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/params"
	"github.com/vinceanalytics/vince/internal/query"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/timeseries"
	"github.com/vinceanalytics/vince/pkg/log"
	"github.com/vinceanalytics/vince/pkg/property"
	"github.com/vinceanalytics/vince/pkg/spec"
)

func Query(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := models.GetSite(ctx)
	var q query.Query
	err := json.NewDecoder(r.Body).Decode(&q)
	if err != nil {
		log.Get().Err(err).Msg("failed to decode query body")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	render.JSON(w, http.StatusOK, timeseries.Query(ctx, site.UserID, site.ID, q))
}

func Delete(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, http.StatusNotImplemented, http.StatusText(http.StatusNotImplemented))
}

func All(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var sid, uid uint64
	if site := models.GetSite(ctx); site != nil {
		sid = site.ID
		uid = site.ID
	} else {
		uid = models.GetUser(ctx).ID
	}
	render.JSON(w, http.StatusOK, timeseries.AllStats(
		ctx, uid, sid,
	))
}

func Metric(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var sid, uid uint64
	if site := models.GetSite(ctx); site != nil {
		sid = site.ID
		uid = site.ID
	} else {
		uid = models.GetUser(ctx).ID
	}
	metric := property.ParsMetric(params.Get(ctx).Get("metric"))
	render.JSON(w, http.StatusOK, timeseries.Stat(
		ctx, uid, sid, metric,
	))
}

func GlobalSeries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	metric := property.ParsMetric(params.Get(ctx).Get("metric"))
	uid := models.GetUser(ctx).ID
	var sid uint64
	if site := models.GetSite(ctx); site != nil {
		sid = site.ID
	}
	var q spec.QueryOptions
	err := json.NewDecoder(r.Body).Decode(&q)
	if err != nil {
		render.JSONError(w, http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}
	render.JSON(w, http.StatusOK, timeseries.QueryGlobalMetric(
		ctx, metric, uid, sid, q.Window, q.Offset,
	))
}
