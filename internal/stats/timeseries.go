package stats

import (
	"net/http"
	"slices"
	"time"

	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/array"
	v1 "github.com/vinceanalytics/vince/gen/go/staples/v1"
	"github.com/vinceanalytics/vince/internal/logger"
	"github.com/vinceanalytics/vince/internal/request"
	"github.com/vinceanalytics/vince/internal/session"
	"github.com/vinceanalytics/vince/internal/timeutil"
)

func TimeSeries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()
	req := v1.Timeseries_Request{
		SiteId:   query.Get("site_id"),
		Period:   ParsePeriod(ctx, query),
		Metrics:  ParseMetrics(ctx, query),
		Interval: ParseInterval(ctx, query),
		Filters:  ParseFilters(ctx, query),
	}
	if !request.Validate(ctx, w, &req) {
		return
	}
	// make sure we have valid interval
	if !ValidByPeriod(req.Period, req.Interval) {
		request.Error(ctx, w, http.StatusBadRequest, "Interval out of range")
		return
	}
	filters := &v1.Filters{
		List: append(req.Filters, &v1.Filter{
			Property: v1.Property_domain,
			Op:       v1.Filter_equal,
			Value:    req.SiteId,
		}),
	}
	metrics := slices.Clone(req.Metrics)
	slices.Sort(metrics)
	metricsToProjection(filters, metrics)
	from, to := PeriodToRange(ctx, time.Now, req.Period, r.URL.Query())
	scanRecord, err := session.Get(ctx).Scan(ctx, from.UnixMilli(), to.UnixMilli(), filters)
	if err != nil {
		logger.Get(ctx).Error("Failed scanning", "err", err)
		request.Internal(ctx, w)
		return
	}

	mapping := map[string]int{}
	for i := 0; i < int(scanRecord.NumCols()); i++ {
		mapping[scanRecord.ColumnName(i)] = i
	}
	tsKey := mapping[v1.Filters_Timestamp.String()]
	ts := scanRecord.Column(tsKey).(*array.Int64).Int64Values()
	var buckets []Bucket
	xc := &Compute{mapping: make(map[string]arrow.Array)}
	err = timeutil.TimeBuckets(req.Interval, ts, func(bucket int64, start, end int) error {
		n := scanRecord.NewSlice(int64(start), int64(end))
		defer n.Release()
		buck := Bucket{
			Timestamp: time.UnixMilli(bucket),
			Values:    make(map[string]float64),
		}
		xc.Reset(n)
		for _, x := range metrics {
			value, err := xc.Metric(ctx, x)
			if err != nil {
				return err
			}
			buck.Values[x.String()] = value
		}
		buckets = append(buckets, buck)
		return nil
	})
	if err != nil {
		logger.Get(ctx).Error("Failed processing buckets", "err", err)
		request.Internal(ctx, w)
		return
	}
	request.Write(ctx, w, &Series{Results: buckets})
}

type Bucket struct {
	Timestamp time.Time          `json:"timestamp"`
	Values    map[string]float64 `json:"values"`
}

type Series struct {
	Results []Bucket `json:"results"`
}
