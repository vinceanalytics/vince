package timeseries

import (
	"bytes"
	"context"
	"encoding/binary"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/query"
	"github.com/vinceanalytics/vince/internal/system"
	"github.com/vinceanalytics/vince/pkg/log"
)

type internalValue struct {
	Timestamp []int64  `json:"timestamp"`
	Value     []uint16 `json:"value"`
}

var (
	defaultStep = time.Hour
)

func Query(ctx context.Context, uid, sid uint64, r query.Query) (result query.QueryResult) {
	currentTime := core.Now(ctx)
	startTS := currentTime

	step := defaultStep

	var window, offset time.Duration
	if r.Window != nil {
		window = r.Window.Value
	}
	if r.Offset != nil {
		offset = r.Offset.Value
	}
	startTS = startTS.Add(-offset)
	endTS := startTS
	startTS = endTS.Add(-window)
	startTS = startTS.Add(time.Second)
	if endTS.Before(startTS) {
		endTS = startTS
	}
	start := startTS.UnixMilli()
	end := endTS.UnixMilli()

	shared := sharedTS(start, end, step.Milliseconds())
	result.Timestamps = shared

	m := newMetaKey()
	defer func() {
		m.Release()
		system.QueryDuration.Observe(time.Since(currentTime).Seconds())
	}()
	m.uid(uid, sid)

	var wg sync.WaitGroup
	now := uint64(currentTime.UnixMilli())
	p := r.Props
	rs := &result.Props
	do(ctx, Base, p.Base, &rs.Base, &wg, shared, now, uint64(start), uint64(end), m.clone())
	do(ctx, Event, p.Event, &rs.Event, &wg, shared, now, uint64(start), uint64(end), m.clone())
	do(ctx, Page, p.Page, &rs.Page, &wg, shared, now, uint64(start), uint64(end), m.clone())
	do(ctx, EntryPage, p.EntryPage, &rs.EntryPage, &wg, shared, now, uint64(start), uint64(end), m.clone())
	do(ctx, ExitPage, p.ExitPage, &rs.ExitPage, &wg, shared, now, uint64(start), uint64(end), m.clone())
	do(ctx, Referrer, p.Referrer, &rs.Referrer, &wg, shared, now, uint64(start), uint64(end), m.clone())
	do(ctx, UtmMedium, p.UtmMedium, &rs.UtmMedium, &wg, shared, now, uint64(start), uint64(end), m.clone())
	do(ctx, UtmSource, p.UtmSource, &rs.UtmSource, &wg, shared, now, uint64(start), uint64(end), m.clone())
	do(ctx, UtmCampaign, p.UtmCampaign, &rs.UtmCampaign, &wg, shared, now, uint64(start), uint64(end), m.clone())
	do(ctx, UtmContent, p.UtmContent, &rs.UtmContent, &wg, shared, now, uint64(start), uint64(end), m.clone())
	do(ctx, UtmTerm, p.UtmTerm, &rs.UtmTerm, &wg, shared, now, uint64(start), uint64(end), m.clone())
	do(ctx, UtmDevice, p.UtmDevice, &rs.UtmDevice, &wg, shared, now, uint64(start), uint64(end), m.clone())
	do(ctx, UtmBrowser, p.UtmBrowser, &rs.UtmBrowser, &wg, shared, now, uint64(start), uint64(end), m.clone())
	do(ctx, BrowserVersion, p.BrowserVersion, &rs.BrowserVersion, &wg, shared, now, uint64(start), uint64(end), m.clone())
	do(ctx, Os, p.Os, &rs.Os, &wg, shared, now, uint64(start), uint64(end), m.clone())
	do(ctx, OsVersion, p.OsVersion, &rs.OsVersion, &wg, shared, now, uint64(start), uint64(end), m.clone())
	do(ctx, Country, p.Country, &rs.Country, &wg, shared, now, uint64(start), uint64(end), m.clone())
	do(ctx, Region, p.Region, &rs.Region, &wg, shared, now, uint64(start), uint64(end), m.clone())
	do(ctx, City, p.City, &rs.City, &wg, shared, now, uint64(start), uint64(end), m.clone())
	wg.Wait()
	return
}

func do(
	ctx context.Context,
	prop Property,
	metrics *query.Metrics,
	result **query.MetricsResult,
	wg *sync.WaitGroup,
	shared []int64,
	now, start, end uint64,
	m *Key,
) {
	if metrics == nil {
		m.Release()
		return
	}
	wg.Add(1)
	go doQuery(ctx, prop, metrics, result, wg, shared, now, start, end, m)
}

func doQuery(
	ctx context.Context,
	prop Property,
	metrics *query.Metrics,
	result **query.MetricsResult,
	wg *sync.WaitGroup,
	shared []int64,
	now, start, end uint64,
	m *Key,
) {
	defer func() {
		m.Release()
		wg.Done()
	}()
	m.prop(prop)
	*result = &query.MetricsResult{}
	r := *result

	txn := GetMike(ctx).NewTransactionAt(now, false)
	o := badger.DefaultIteratorOptions
	o.SinceTs = uint64(start)
	// make sure all iterations are in /user_id/site_id/ scope
	o.Prefix = m[:metricOffset]
	it := txn.NewIterator(o)
	defer it.Close()
	b := get()
	defer put(b)
	getMetric(Visitors, end, shared, it, b, m, metrics.Visitors, &r.Visitors)
	getMetric(Views, end, shared, it, b, m, metrics.Visitors, &r.Views)
	getMetric(Events, end, shared, it, b, m, metrics.Visitors, &r.Events)
	getMetric(Visits, end, shared, it, b, m, metrics.Visitors, &r.Visits)
	getMetric(BounceRates, end, shared, it, b, m, metrics.Visitors, &r.BounceRates)
	getMetric(VisitDurations, end, shared, it, b, m, metrics.Visitors, &r.VisitDurations)
}

func getMetric(
	metric Metric,
	end uint64,
	shared []int64,
	it *badger.Iterator,
	b *bytes.Buffer,
	m *Key,
	sel *query.Select,
	result *map[string][]uint32,
) {
	if sel == nil {
		return
	}
	*result = make(map[string][]uint32)
	var text string
	if sel.Exact != nil {
		text = sel.Exact.Value
	}
	b.Reset()
	values := make(map[string]*internalValue)

	prefix := m.metric(metric).idx(b, text).Bytes()
	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		x := it.Item()
		if x.Version() > end {
			// We have reached the end of iteration. Range actually
			// reflects on the version we are interested in.
			break
		}
		kb := x.Key()
		// text comes after the key offset
		txt := kb[keyOffset : len(kb)-6]
		if !sel.Match(txt) {
			continue
		}
		v, ok := values[string(txt)]
		if !ok {
			v = &internalValue{}
			values[string(txt)] = v
		}
		err := x.Value(func(val []byte) error {
			v.Timestamp = append(v.Timestamp, int64(x.Version()))
			v.Value = append(v.Value,
				binary.BigEndian.Uint16(val),
			)
			return nil
		})
		if err != nil {
			log.Get().Err(err).Msg("failed to read value from kv store")
		}
	}
	o := *result
	for k, v := range values {
		o[k] = rollUp(v.Value, v.Timestamp, shared, Sum16)
	}
}

func Sum(ls []uint32) (o uint32) {
	for _, v := range ls {
		o += v
	}
	return
}

func Sum16(ls []uint16) (o uint32) {
	for _, v := range ls {
		o += uint32(v)
	}
	return
}
