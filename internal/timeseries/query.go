package timeseries

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"path"
	"regexp"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/pkg/log"
	"github.com/vinceanalytics/vince/pkg/property"
	"github.com/vinceanalytics/vince/pkg/spec"
	"github.com/vinceanalytics/vince/pkg/timex"
)

func QuerySeries(ctx context.Context, uid, sid uint64, o spec.QueryPropertyOptions) (result spec.PropertyResult[[]uint64]) {
	return queryProperty[[]uint64](ctx, uid, sid, o)
}

func QueryAggregate(ctx context.Context, uid, sid uint64, o spec.QueryPropertyOptions) (result spec.PropertyResult[uint64]) {
	return queryProperty[uint64](ctx, uid, sid, o)
}

func queryProperty[T uint64 | []uint64](ctx context.Context, uid, sid uint64, o spec.QueryPropertyOptions) (result spec.PropertyResult[T]) {
	now := core.Now(ctx)
	sel := selector(o.Selector)
	if sel.invalid {
		return
	}
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

func Stat(ctx context.Context, uid, sid uint64, metric property.Metric) spec.Global[uint64] {
	return global[uint64](ctx, uid, sid, metric)
}

func Stats(ctx context.Context, uid, sid uint64) spec.Global[spec.Metrics] {
	return global[spec.Metrics](ctx, uid, sid, property.Metric(0))
}

func global[T uint64 | spec.Metrics](ctx context.Context, uid, sid uint64, metric property.Metric) (o spec.Global[T]) {
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
			u64(txn, key, property.Visitors, &e.Visitors),
			u64(txn, key, property.Views, &e.Views),
			u64(txn, key, property.Events, &e.Events),
			u64(txn, key, property.Visits, &e.Visits),
		)
	}
	if err != nil {
		log.Get().Err(err).Msg("failed to query global stats")
	}
	txn.Discard()
	m.Release()
	put(b)
	o.Elapsed = core.Elapsed(ctx, start)
	return
}

func GlobalAggregate(ctx context.Context, uid, sid uint64, o spec.QueryOptions) (r spec.Series[uint64]) {
	return queryGlobal[uint64](ctx, uid, sid, o)
}

func GlobalSeries(ctx context.Context, uid, sid uint64, o spec.QueryOptions) (r spec.Series[[]uint64]) {
	return queryGlobal[[]uint64](ctx, uid, sid, o)
}

func queryGlobal[T uint64 | []uint64](ctx context.Context, uid, sid uint64, o spec.QueryOptions) (r spec.Series[T]) {
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
	put(b)
	switch e := any(&r.Result).(type) {
	case *uint64:
		*e = Sum64(values)
	case *[]uint64:
		*e = rollUp(values, ts, r.Timestamps, Sum64)
	}
	r.Elapsed = core.Elapsed(ctx, now)
	return
}

func u64(txn *badger.Txn, b []byte, m property.Metric, o *uint64) error {
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

type internalValue struct {
	Timestamp []int64  `json:"timestamp"`
	Value     []uint64 `json:"value"`
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
