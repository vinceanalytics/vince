package timeseries

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"math"
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
	if r.Sum {
		// for sum only use single timestamp
		result.Timestamps = []int64{shared[len(shared)-1]}
	}

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
	do(ctx, Base, p.Base, &rs.Base, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	do(ctx, Event, p.Event, &rs.Event, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	do(ctx, Page, p.Page, &rs.Page, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	do(ctx, EntryPage, p.EntryPage, &rs.EntryPage, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	do(ctx, ExitPage, p.ExitPage, &rs.ExitPage, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	do(ctx, Referrer, p.Referrer, &rs.Referrer, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	do(ctx, UtmMedium, p.UtmMedium, &rs.UtmMedium, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	do(ctx, UtmSource, p.UtmSource, &rs.UtmSource, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	do(ctx, UtmCampaign, p.UtmCampaign, &rs.UtmCampaign, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	do(ctx, UtmContent, p.UtmContent, &rs.UtmContent, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	do(ctx, UtmTerm, p.UtmTerm, &rs.UtmTerm, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	do(ctx, UtmDevice, p.UtmDevice, &rs.UtmDevice, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	do(ctx, UtmBrowser, p.UtmBrowser, &rs.UtmBrowser, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	do(ctx, BrowserVersion, p.BrowserVersion, &rs.BrowserVersion, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	do(ctx, Os, p.Os, &rs.Os, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	do(ctx, OsVersion, p.OsVersion, &rs.OsVersion, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	do(ctx, Country, p.Country, &rs.Country, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	do(ctx, Region, p.Region, &rs.Region, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	do(ctx, City, p.City, &rs.City, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
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
	sum bool,
) {
	if metrics == nil {
		m.Release()
		return
	}
	wg.Add(1)
	go doQuery(ctx, prop, metrics, result, wg, shared, now, start, end, m, sum)
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
	sum bool,
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
	getMetric(sum, Visitors, metrics.Visitors, end, shared, it, b, m, &r.Visitors)
	getMetric(sum, Views, metrics.Views, end, shared, it, b, m, &r.Views)
	getMetric(sum, Events, metrics.Events, end, shared, it, b, m, &r.Events)
	getMetric(sum, Visits, metrics.Visits, end, shared, it, b, m, &r.Visits)
	// bounce rate is a percentage of bounce to visits. We only save bounce counts,
	// so we must calculate the rate here.
	getMetric(sum, BounceRates, metrics.BounceRates, end, shared, it, b, m, &r.BounceRates)
	if metrics.BounceRates != nil {
		o := r.Visits
		if metrics.Visits == nil || !metrics.Visits.Equal(metrics.BounceRates) {
			o = make(map[string][]uint32)
			getMetric(sum, Visits, metrics.BounceRates, end, shared, it, b, m, &o)
		}
		for k, v := range r.BounceRates {
			xv, ok := o[k]
			if !ok {
				continue
			}
			percent(v, xv)
		}
	}
	getMetric(sum, VisitDurations, metrics.VisitDurations, end, shared, it, b, m, &r.VisitDurations)
}

func percent(a, b []uint32) {
	for i := range a {
		if b[i] == 0 || a[i] == 0 {
			// avoid dividing by zero
			continue
		}
		x := float64(a[i]) / float64(b[i])
		x = math.Round(x)
		a[i] = uint32(x)
	}
}
func getMetric(
	sum bool,
	metric Metric,
	sel *query.Select,
	end uint64,
	shared []int64,
	it *badger.Iterator,
	b *bytes.Buffer,
	m *Key,
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
		if sum {
			o[k] = []uint32{Sum16(v.Value)}
		} else {
			o[k] = rollUp(v.Value, v.Timestamp, shared, Sum16)
		}
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

func Global(ctx context.Context, uid, sid uint64) (o query.Global) {
	now := core.Now(ctx).UnixMilli()
	txn := GetMike(ctx).NewTransactionAt(uint64(now), false)
	m := newMetaKey()
	m.uid(uid, sid)
	b := get()
	err := errors.Join(
		u16(txn, m.metric(Visitors).site(b), &o.Visitors),
		u16(txn, m.metric(Views).site(b), &o.Views),
		u16(txn, m.metric(Events).site(b), &o.Events),
		u16(txn, m.metric(Visits).site(b), &o.Visits),
	)
	if err != nil {
		log.Get().Err(err).Msg("failed to query global stats")
	}
	return
}

func u16(txn *badger.Txn, key *bytes.Buffer, o *uint64) error {
	it, err := txn.Get(key.Bytes())
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil
		}
	}
	return it.Value(func(val []byte) error {
		*o = binary.BigEndian.Uint64(val)
		return nil
	})
}
