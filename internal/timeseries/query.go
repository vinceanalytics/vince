package timeseries

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"math"
	"path"
	"regexp"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/query"
	"github.com/vinceanalytics/vince/internal/system"
	"github.com/vinceanalytics/vince/pkg/log"
	"github.com/vinceanalytics/vince/pkg/property"
	"github.com/vinceanalytics/vince/pkg/spec"
	"github.com/vinceanalytics/vince/pkg/timex"
)

type internalValue struct {
	Timestamp []int64  `json:"timestamp"`
	Value     []uint64 `json:"value"`
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

	start := startTS.Truncate(time.Hour).UnixMilli()
	end := endTS.Truncate(time.Hour).UnixMilli()

	shared := sharedTS(start, end, step.Milliseconds())

	result.Timestamps = shared
	if r.Sum {
		// for sum only use single timestamp
		result.Timestamps = []int64{shared[len(shared)-1]}
	}

	m := newMetaKey()
	defer func() {
		m.Release()
		system.QueryDuration.Observe(core.Now(ctx).Sub(currentTime).Seconds())
	}()
	m.uid(uid, sid)

	var wg sync.WaitGroup
	now := uint64(currentTime.UnixMilli())
	p := r.Props
	rs := &result.Props
	if p.Event != nil {
		do(ctx, Event, p.Event, &rs.Event, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	}
	if p.Page != nil {
		do(ctx, Page, p.Page, &rs.Page, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	}
	if p.EntryPage != nil {
		do(ctx, EntryPage, p.EntryPage, &rs.EntryPage, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	}
	if p.ExitPage != nil {
		do(ctx, ExitPage, p.ExitPage, &rs.ExitPage, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	}
	if p.Referrer != nil {
		do(ctx, Referrer, p.Referrer, &rs.Referrer, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	}
	if p.UtmMedium != nil {
		do(ctx, UtmMedium, p.UtmMedium, &rs.UtmMedium, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	}
	if p.UtmSource != nil {
		do(ctx, UtmSource, p.UtmSource, &rs.UtmSource, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	}
	if p.UtmCampaign != nil {
		do(ctx, UtmCampaign, p.UtmCampaign, &rs.UtmCampaign, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	}
	if p.UtmContent != nil {
		do(ctx, UtmContent, p.UtmContent, &rs.UtmContent, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	}
	if p.UtmTerm != nil {
		do(ctx, UtmTerm, p.UtmTerm, &rs.UtmTerm, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	}
	if p.UtmDevice != nil {
		do(ctx, UtmDevice, p.UtmDevice, &rs.UtmDevice, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	}
	if p.UtmBrowser != nil {
		do(ctx, UtmBrowser, p.UtmBrowser, &rs.UtmBrowser, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	}
	if p.BrowserVersion != nil {
		do(ctx, BrowserVersion, p.BrowserVersion, &rs.BrowserVersion, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	}
	if p.Os != nil {
		do(ctx, Os, p.Os, &rs.Os, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	}
	if p.OsVersion != nil {
		do(ctx, OsVersion, p.OsVersion, &rs.OsVersion, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	}
	if p.Country != nil {
		do(ctx, Country, p.Country, &rs.Country, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	}
	if p.Region != nil {
		do(ctx, Region, p.Region, &rs.Region, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	}
	if p.City != nil {
		do(ctx, City, p.City, &rs.City, &wg, shared, now, uint64(start), uint64(end), m.clone(), r.Sum)
	}
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

	txn := Permanent(ctx).NewTransactionAt(now, false)
	o := badger.DefaultIteratorOptions
	o.SinceTs = uint64(start)
	// make sure all iterations are in /user_id/site_id/ scope
	o.Prefix = m[:metricOffset]
	it := txn.NewIterator(o)
	defer it.Close()
	b := get()
	defer put(b)
	if metrics.Visitors != nil {
		getMetric(sum, Visitors, metrics.Visitors, end, shared, it, b, m, &r.Visitors)
	}
	if metrics.Views != nil {
		getMetric(sum, Views, metrics.Views, end, shared, it, b, m, &r.Views)
	}
	if metrics.Events != nil {
		getMetric(sum, Events, metrics.Events, end, shared, it, b, m, &r.Events)
	}
	if metrics.Visits != nil {
		getMetric(sum, Visits, metrics.Visits, end, shared, it, b, m, &r.Visits)
	}
	if metrics.BounceRates != nil {
		getMetric(sum, BounceRates, metrics.BounceRates, end, shared, it, b, m, &r.BounceRates)
		visits := r.Visits
		if metrics.Visits == nil || !metrics.Visits.Equal(metrics.BounceRates) {
			visits = make(map[string][]uint64)
			getMetric(sum, Visits, metrics.BounceRates, end, shared, it, b, m, &visits)
		}
		for k, v := range r.BounceRates {
			xv, ok := visits[k]
			if !ok {
				continue
			}
			percent(v, xv)
		}
	}
	if metrics.VisitDurations != nil {
		getMetric(sum, VisitDurations, metrics.VisitDurations, end, shared, it, b, m, &r.VisitDurations)
		visits := r.Visits
		if metrics.Visits == nil || !metrics.Visits.Equal(metrics.VisitDurations) {
			visits = make(map[string][]uint64)
			getMetric(sum, Visits, metrics.VisitDurations, end, shared, it, b, m, &visits)
		}
		for k, v := range r.VisitDurations {
			xv, ok := visits[k]
			if !ok {
				continue
			}
			percent(v, xv)
		}
	}
}

func percent(a, b []uint64) {
	for i := range a {
		if b[i] == 0 || a[i] == 0 {
			// avoid dividing by zero
			a[i] = 0
			continue
		}
		x := float64(a[i]) / float64(b[i])
		x = math.Round(x)
		a[i] = uint64(x)
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
	result *map[string][]uint64,
) {
	*result = make(map[string][]uint64)
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
		ts := binary.BigEndian.Uint64(kb[len(kb)-8:])
		if ts == 0 || ts > end {
			continue
		}
		txt := kb[:len(kb)-8]
		if !sel.Match(txt) {
			continue
		}
		v, ok := values[string(txt)]
		if !ok {
			v = &internalValue{}
			values[string(txt)] = v
		}
		err := x.Value(func(val []byte) error {
			v.Timestamp = append(v.Timestamp, int64(ts))
			v.Value = append(v.Value,
				binary.BigEndian.Uint64(val),
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
			o[k] = []uint64{Sum64(v.Value)}
		} else {
			o[k] = rollUp(v.Value, v.Timestamp, shared, Sum64)
		}
	}
}

func Sum64(ls []uint64) (o uint64) {
	for _, v := range ls {
		o += v
	}
	return
}

type sel struct {
	exact, glob string
	invalid     bool
	re          *regexp.Regexp
}

func (s *sel) Match(txt []byte) bool {
	if s.exact != "" {
		return s.exact == string(txt)
	}
	if s.glob != "" {
		ok, _ := path.Match(s.glob, string(txt))
		return ok
	}
	if s.re != nil {
		return s.re.Match(txt)
	}
	return true
}

func selector(o spec.Select) (s sel) {
	if o.Exact != nil {
		s.exact = *o.Exact
	}
	if o.Glob != nil {
		s.glob = *o.Glob
	}
	if o.Re != nil {
		var err error
		s.re, err = regexp.Compile(*o.Re)
		if err != nil {
			log.Get().Err(err).
				Str("pattern", *o.Re).
				Msg("failed to compile regular expression for selector")
			s.invalid = true
		}
	}
	return
}

func QuerySeries(ctx context.Context, uid, sid uint64, o spec.QueryPropertyOptions) (result spec.PropertyResult[[]uint64]) {
	return queryProperty[[]uint64](ctx, uid, sid, o)
}

func QueryAggregate(ctx context.Context, uid, sid uint64, o spec.QueryPropertyOptions) (result spec.PropertyResult[uint64]) {
	return queryProperty[uint64](ctx, uid, sid, o)
}

func queryProperty[T uint64 | []uint64](ctx context.Context, uid, sid uint64, o spec.QueryPropertyOptions) (result spec.PropertyResult[T]) {
	now := core.Now(ctx)
	window := o.Window
	if window < time.Hour {
		window = timex.Today.Window(now)
	}

	start := now.Add(-o.Offset)
	end := start
	start = end.Add(-window)
	if end.Before(start) {
		end = start
	}
	switch any(result.Result).(type) {
	case map[string][]uint64:
		result.Timestamps = sharedTS(start.UnixMilli(), end.UnixMilli(), time.Hour.Milliseconds())
	}
	result.Result = make(map[string]T)

	readTs := uint64(now.UnixMilli())
	startTs := uint64(start.UnixMilli())
	endTs := uint64(end.UnixMilli())

	txn := Permanent(ctx).NewTransactionAt(readTs, false)
	m := newMetaKey()
	m.uid(uid, sid)
	m.metric(o.Metric).prop(o.Property)

	b := get()
	b.Write(m[:])
	var text string
	sel := selector(o.Selector)
	if sel.exact != "" {
		text = sel.exact
	}
	b.WriteString(text)
	prefix := b.Bytes()

	opt := badger.DefaultIteratorOptions
	opt.Prefix = prefix
	opt.SinceTs = startTs

	values := make(map[string]*internalValue)

	it := txn.NewIterator(opt)
	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		item := it.Item()
		if item.Version() >= endTs {
			break
		}
		key := item.Key()
		txt := key[len(m) : len(key)-8]
		if !sel.Match(txt) {
			continue
		}
		x, ok := values[string(txt)]
		if !ok {
			x = &internalValue{}
			values[string(txt)] = x
		}
		item.Value(func(val []byte) error {
			x.Timestamp = append(x.Timestamp, int64(item.Version()))
			x.Value = append(x.Value,
				binary.BigEndian.Uint64(val),
			)
			return nil
		})
	}
	it.Close()
	m.Release()
	put(b)

	switch r := any(result.Result).(type) {
	case map[string][]uint64:
		for k, v := range values {
			r[k] = rollUp(v.Value, v.Timestamp, result.Timestamps, Sum64)
		}
	case map[string]uint64:
		for k, v := range values {
			r[k] = Sum64(v.Value)
		}
	}
	result.Elapsed = core.Elapsed(ctx, now)
	return
}

func Stat(ctx context.Context, uid, sid uint64, metric Metric) spec.Global[uint64] {
	return global[uint64](ctx, uid, sid, metric)
}

func AllStats(ctx context.Context, uid, sid uint64) spec.Global[spec.Metrics] {
	return global[spec.Metrics](ctx, uid, sid, property.Metric(0))
}

func global[T uint64 | spec.Metrics](ctx context.Context, uid, sid uint64, metric Metric) (o spec.Global[T]) {
	start := core.Now(ctx)
	now := start.UnixMilli()
	txn := Permanent(ctx).NewTransactionAt(uint64(now), false)
	m := newMetaKey()
	m.uid(uid, sid)
	b := get()

	b.Write(m[:])
	b.Write(zero)
	key := b.Bytes()
	var err error
	switch e := any(&o.Result).(type) {
	case *uint64:
		err = u64(txn, key, metric, e)
	case *spec.Metrics:
		err = errors.Join(
			u64(txn, key, Visitors, &e.Visitors),
			u64(txn, key, Views, &e.Views),
			u64(txn, key, Events, &e.Events),
			u64(txn, key, Visits, &e.Visits),
		)
	}
	if err != nil {
		log.Get().Err(err).Msg("failed to query global stats")
	}
	m.Release()
	put(b)
	o.Elapsed = core.Elapsed(ctx, start)
	return
}

func GlobalAggregate(ctx context.Context, uid, sid uint64, o spec.QueryOptions) (r spec.ResultSet[uint64]) {
	return queryGlobal[uint64](ctx, uid, sid, o)
}

func GlobalSeries(ctx context.Context, uid, sid uint64, o spec.QueryOptions) (r spec.ResultSet[[]uint64]) {
	return queryGlobal[[]uint64](ctx, uid, sid, o)
}

func queryGlobal[T uint64 | []uint64](ctx context.Context, uid, sid uint64, o spec.QueryOptions) (r spec.ResultSet[T]) {
	now := core.Now(ctx)
	start := now.Add(-o.Offset)
	end := start
	start = end.Add(-o.Window)
	if end.Before(start) {
		end = start
	}
	readTs := uint64(now.UnixMilli())
	endTs := uint64(end.UnixMilli())

	var ts []int64
	var values []uint64
	txn := Permanent(ctx).NewTransactionAt(readTs, false)
	m := newMetaKey()
	b := get()
	m.uid(uid, sid)
	m.metric(o.Metric)
	b.Write(m[:])

	opts := badger.DefaultIteratorOptions
	switch any(r.Result).(type) {
	case []uint64:
		r.Timestamps = sharedTS(start.UnixMilli(), end.UnixMilli(), time.Hour.Milliseconds())
	}
	prefix := m[:]

	opts.Prefix = prefix
	it := txn.NewIterator(opts)
	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		item := it.Item()
		if item.Version() >= endTs {
			break
		}
		k := item.Key()
		if bytes.Equal(k[len(k)-8:], zero) {
			continue
		}
		item.Value(func(val []byte) error {
			ts = append(ts, int64(item.Version()))
			values = append(values, binary.BigEndian.Uint64(val))
			return nil
		})
	}
	it.Close()
	m.Release()
	switch e := any(&r.Result).(type) {
	case *uint64:
		*e = Sum64(values)
	case *[]uint64:
		*e = rollUp(values, ts, r.Timestamps, Sum64)
	}
	r.Elapsed = core.Elapsed(ctx, now)
	return
}

func u64(txn *badger.Txn, b []byte, m Metric, o *uint64) error {
	b[metricOffset] = byte(m)
	it, err := txn.Get(b)
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
