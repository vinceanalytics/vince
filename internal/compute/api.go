package compute

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/apache/arrow/go/v15/arrow/compute"
	"github.com/bufbuild/protovalidate-go"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/columns"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/logger"
	"github.com/vinceanalytics/vince/internal/timeutil"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var validate *protovalidate.Validator

func init() {
	var err error
	validate, err = protovalidate.New(protovalidate.WithFailFast(true))
	if err != nil {
		logger.Fail("Failed setting up validator", "err", err)
	}
}

func Realtime(ctx context.Context, scan db.Scanner, req *v1.Realtime_Request) (*v1.Realtime_Response, error) {
	err := validate.Validate(req)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	firstTime := now.Add(-5 * time.Minute)
	result, err := scan.Scan(ctx,
		req.TenantId,
		firstTime.UnixMilli(),
		now.UnixMilli(),
		&v1.Filters{
			Projection: []v1.Filters_Projection{
				v1.Filters_id,
			},
			List: []*v1.Filter{
				{Property: v1.Property_domain, Op: v1.Filter_equal, Value: req.SiteId},
			},
		},
	)
	if err != nil {
		return nil, err
	}
	defer result.Release()
	m := NewCompute(result)
	visitors, err := m.Visitors(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.Realtime_Response{Visitors: uint64(visitors)}, nil
}

func Aggregate(ctx context.Context, scan db.Scanner, req *v1.Aggregate_Request) (*v1.Aggregate_Response, error) {
	err := validate.Validate(req)
	if err != nil {
		return nil, err
	}
	filters := &v1.Filters{
		List: append(req.Filters, &v1.Filter{
			Property: v1.Property_domain,
			Op:       v1.Filter_equal,
			Value:    req.SiteId,
		}),
	}
	slices.Sort(req.Metrics)
	MetricsToProjection(filters, req.Metrics)
	from, to := periodToRange(time.Now, req.Period, req.Date)
	resultRecord, err := scan.Scan(ctx, req.TenantId, from.UnixMilli(), to.UnixMilli(), filters)
	if err != nil {
		return nil, err
	}
	defer resultRecord.Release()
	mapping := map[string]arrow.Array{}
	for i := 0; i < int(resultRecord.NumCols()); i++ {
		mapping[resultRecord.ColumnName(i)] = resultRecord.Column(i)
	}
	res := &v1.Aggregate_Response{
		Results: make(map[string]float64),
	}
	xc := &Compute{Mapping: mapping}
	for _, metric := range req.Metrics {
		value, err := xc.Metric(ctx, metric)
		if err != nil {
			return nil, err
		}
		res.Results[metric.String()] = value
	}
	return res, nil
}

func Breakdown(ctx context.Context, scan db.Scanner, req *v1.BreakDown_Request) (*v1.BreakDown_Response, error) {
	err := validate.Validate(req)
	if err != nil {
		return nil, err
	}
	filter := &v1.Filters{
		List: append(req.Filters, &v1.Filter{
			Property: v1.Property_domain,
			Op:       v1.Filter_equal,
			Value:    req.SiteId,
		}),
	}
	slices.Sort(req.Metrics)
	slices.Sort(req.Property)
	selectedColumns := MetricsToProjection(filter, req.Metrics, req.Property...)
	from, to := periodToRange(time.Now, req.Period, req.Date)
	scannedRecord, err := scan.Scan(ctx, req.TenantId, from.UnixMilli(), to.UnixMilli(), filter)
	if err != nil {
		return nil, err
	}
	defer scannedRecord.Release()
	mapping := map[string]arrow.Array{}
	for i := 0; i < int(scannedRecord.NumCols()); i++ {
		mapping[scannedRecord.ColumnName(i)] = scannedRecord.Column(i)
	}
	defer clear(mapping)
	b := array.NewUint32Builder(compute.GetAllocator(ctx))
	defer b.Release()
	// TODO: run this concurrently
	xc := &Compute{
		Mapping: make(map[string]arrow.Array),
	}
	defer xc.Release()
	res := &v1.BreakDown_Response{}
	for _, prop := range req.Property {
		rp := &v1.BreakDown_Result{
			Property: prop,
		}
		for key, bitmap := range HashProp(mapping[prop.String()]) {
			b.AppendValues(bitmap.ToArray(), nil)
			idx := b.NewUint32Array()

			kv := &v1.BreakDown_KeyValues{
				Key:   key,
				Value: make(map[string]float64),
			}
			for _, name := range selectedColumns {
				a, err := Take(ctx, mapping[name], idx)
				if err != nil {
					return nil, err
				}
				xc.Mapping[name] = a
			}
			for _, metric := range req.Metrics {
				value, err := xc.Metric(ctx, metric)
				if err != nil {
					return nil, err
				}
				kv.Value[metric.String()] = value
			}
			rp.Values = append(rp.Values, kv)
			xc.Release()
		}
		res.Results = append(res.Results, rp)

	}
	return res, nil
}

func Timeseries(ctx context.Context, scan db.Scanner, req *v1.Timeseries_Request) (*v1.Timeseries_Response, error) {
	err := validate.Validate(req)
	if err != nil {
		return nil, err
	}
	if !ValidByPeriod(req.Period, req.Interval) {
		return nil, errors.New("invalid interval")
	}
	filters := &v1.Filters{
		List: append(req.Filters, &v1.Filter{
			Property: v1.Property_domain,
			Op:       v1.Filter_equal,
			Value:    req.SiteId,
		}),
	}
	slices.Sort(req.Metrics)
	MetricsToProjection(filters, req.Metrics)
	from, to := periodToRange(time.Now, req.Period, req.Date)
	scanRecord, err := scan.Scan(ctx, req.TenantId, from.UnixMilli(), to.UnixMilli(), filters)
	if err != nil {
		return nil, err
	}
	defer scanRecord.Release()

	mapping := map[string]int{}
	for i := 0; i < int(scanRecord.NumCols()); i++ {
		mapping[scanRecord.ColumnName(i)] = i
	}
	tsKey := mapping[columns.Timestamp]
	ts := scanRecord.Column(tsKey).(*array.Int64).Int64Values()
	var buckets []*v1.Timeseries_Bucket
	xc := &Compute{Mapping: make(map[string]arrow.Array)}
	defer xc.Reset(nil)
	err = timeutil.TimeBuckets(req.Interval, ts, func(bucket int64, start, end int) error {
		n := scanRecord.NewSlice(int64(start), int64(end))
		defer n.Release()
		buck := &v1.Timeseries_Bucket{
			Timestamp: timestamppb.New(time.UnixMilli(bucket)),
			Values:    make(map[string]float64),
		}
		xc.Reset(n)
		for _, x := range req.Metrics {
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
		return nil, err
	}
	return &v1.Timeseries_Response{Results: buckets}, nil
}
func periodToRange(now nowFunc, period *v1.TimePeriod, tsDate *timestamppb.Timestamp) (start, end time.Time) {
	date := parseDate(now, tsDate)
	switch e := period.Value.(type) {
	case *v1.TimePeriod_Base_:
		switch e.Base {
		case v1.TimePeriod_day:
			end = date
			start = timeutil.BeginDay(end)
		case v1.TimePeriod__7d:
			end = date
			start = end.AddDate(0, 0, -6)
		case v1.TimePeriod__30d:
			end = date
			start = end.AddDate(0, 0, -30)
		case v1.TimePeriod_mo:
			end = date
			start = timeutil.BeginMonth(end)
			end = timeutil.EndMonth(end)
		case v1.TimePeriod__6mo:
			end = timeutil.EndMonth(date)
			start = timeutil.BeginMonth(end.AddDate(0, -5, 0))
		case v1.TimePeriod__12mo:
			end = timeutil.EndMonth(date)
			start = timeutil.BeginMonth(end.AddDate(0, -11, 0))
		case v1.TimePeriod_year:
			end = timeutil.EndYear(date)
			start = timeutil.BeginYear(end)
		}

	case *v1.TimePeriod_Custom_:
		end = e.Custom.End.AsTime()
		start = e.Custom.Start.AsTime()
	}
	return
}

type nowFunc func() time.Time

func parseDate(now nowFunc, ts *timestamppb.Timestamp) time.Time {
	if ts != nil {
		return ts.AsTime()
	}
	return timeutil.EndDay(now())
}

func ValidByPeriod(period *v1.TimePeriod, i v1.Interval) bool {
	switch e := period.Value.(type) {
	case *v1.TimePeriod_Base_:
		switch e.Base {
		case v1.TimePeriod_day:
			return i == v1.Interval_minute || i == v1.Interval_hour
		case v1.TimePeriod__7d:
			return i == v1.Interval_hour || i == v1.Interval_date
		case v1.TimePeriod_mo, v1.TimePeriod__30d:
			return i == v1.Interval_date || i == v1.Interval_week
		case v1.TimePeriod__6mo, v1.TimePeriod__12mo, v1.TimePeriod_year:
			return i == v1.Interval_date || i == v1.Interval_week || i == v1.Interval_month
		default:
			return false
		}
	case *v1.TimePeriod_Custom_:
		return i == v1.Interval_date || i == v1.Interval_week || i == v1.Interval_month
	default:
		return false
	}
}
