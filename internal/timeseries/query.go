package timeseries

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"path"
	"regexp"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/query"
	"github.com/vinceanalytics/vince/internal/system"
	"github.com/vinceanalytics/vince/pkg/log"
)

type BaseQuery struct {
	Start   time.Time     `json:"start,omitempty"`
	Offset  time.Duration `json:"offset,omitempty"`
	Step    time.Duration `json:"step,omitempty"`
	Window  time.Duration `json:"window,omitempty"`
	Metrics []Metric      `json:"metrics"`
	Filters FilterList    `json:"filters"`
}

type QueryRequest struct {
	UserID uint64
	SiteID uint64
	BaseQuery
}

type QueryResult struct {
	ELapsed    time.Duration    `json:"elapsed"`
	Start      time.Time        `json:"start"`
	End        time.Time        `json:"end"`
	Timestamps []int64          `json:"timestamps"`
	Result     PropertiesResult `json:"result"`
}

type Filter struct {
	Property Property   `json:"prop"`
	Omit     []Metric   `json:"omitMetrics,omitempty"`
	Expr     FilterExpr `json:"expr"`
}

func (f *Filter) Accept(m Metric) bool {
	for _, v := range f.Omit {
		if v == m {
			return false
		}
	}
	return true
}

type FilterList []*Filter

func (f FilterList) Validate() error {
	if len(f) == 0 {
		return errors.New("empty filters")
	}
	for _, v := range f {
		err := v.Expr.Validate()
		if err != nil {
			return fmt.Errorf("%s: %v", v.Property, err)
		}
	}
	return nil
}

type FilterExpr struct {
	Property Property       `json:"prop"`
	And      FilterAnd      `json:"and,omitempty"`
	Or       FilterAnd      `json:"or,omitempty"`
	Text     string         `json:"text,omitempty"`
	IsRe     bool           `json:"isRe,omitempty"`
	IsGlob   bool           `json:"isGlob,omitempty"`
	Re       *regexp.Regexp `json:"-"`
}

func (m *FilterExpr) ExactMatch() bool {
	return m.Text != "" &&
		// * is a special case where we match any key. This is an optimization to
		// avoid glob or compiling regular expression.
		m.Text != "*" &&
		len(m.And) == 0 && len(m.Or) == 0 && !m.IsRe && !m.IsGlob
}

type FilterAnd []*FilterExpr

func (f FilterAnd) Match(txt []byte) bool {
	for _, v := range f {
		if !v.Match(txt) {
			return false
		}
	}
	return true
}

type FilterOr []*FilterExpr

func (f FilterOr) Match(txt []byte) bool {
	for _, v := range f {
		if v.Match(txt) {
			return true
		}
	}
	return false
}

func (m *FilterExpr) Validate() error {
	err := m.Compile()
	if err != nil {
		return err
	}
	if len(m.And) > 0 && len(m.Or) > 0 {
		return errors.New("and & or cannot exists in the ame expression")
	}
	empty := len(m.And) == 0 && len(m.Or) == 0 && !m.IsRe && !m.IsGlob && m.Text == ""
	if empty {
		return errors.New("filter expression cannot be empty string")
	}
	return nil
}

func (m *FilterExpr) Compile() error {
	for _, v := range m.And {
		err := v.Compile()
		if err != nil {
			return err
		}
	}
	for _, v := range m.Or {
		err := v.Compile()
		if err != nil {
			return err
		}
	}
	if m.IsRe {
		re, err := regexp.Compile(m.Text)
		if err != nil {
			return fmt.Errorf("bad regular expression %q", m.Text)
		}
		m.Re = re
	}
	return nil
}

func (m *FilterExpr) Match(txt []byte) bool {
	if m.And != nil {
		return m.And.Match(txt)
	}
	if m.Or != nil {
		return m.Or.Match(txt)
	}
	if m.IsRe {
		return m.Re.Match(txt)
	}
	if m.IsGlob {
		ok, _ := path.Match(m.Text, string(txt))
		return ok
	}
	if m.Text == "*" {
		// special case for matching everything
		return true
	}
	return bytes.Equal([]byte(m.Text), txt)
}

type Value struct {
	Timestamp []int64   `json:"timestamp"`
	Value     []float64 `json:"value"`
}

type internalValue struct {
	Timestamp []int64  `json:"timestamp"`
	Value     []uint16 `json:"value"`
}

var (
	defaultStep = time.Hour
)

type PropertiesResult map[string]MetricResult

type MetricResult map[string]OutValue

type OutValue map[string][]uint32

func Query(ctx context.Context, r QueryRequest) (result QueryResult) {
	currentTime := core.Now(ctx)
	startTS := currentTime
	if !r.Start.IsZero() {
		startTS = r.Start
	}

	step := defaultStep
	if r.Step > 0 {
		step = r.Step
	}

	var window, offset time.Duration
	if r.Window > 0 {
		window = r.Window
	}
	if r.Offset > 0 {
		offset = r.Offset
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
	result.Start, result.End = startTS, endTS

	shared := sharedTS(start, end, step.Milliseconds())
	result.Timestamps = shared
	result.Result = make(PropertiesResult)

	txn := GetMike(ctx).NewTransactionAt(uint64(currentTime.UnixMilli()), false)

	m := newMetaKey()
	defer func() {
		m.Release()
		txn.Discard()
		result.ELapsed = time.Since(currentTime)
		system.QueryDuration.Observe(result.ELapsed.Seconds())
	}()
	m.uid(r.UserID, r.SiteID)
	if len(r.Metrics) == 0 {
		return
	}
	b := smallBufferpool.Get().(*bytes.Buffer)
	defer func() {
		b.Reset()
		smallBufferpool.Put(b)
	}()
	o := badger.DefaultIteratorOptions
	o.SinceTs = uint64(start)

	// make sure all iterations are in /user_id/site_id/ scope
	o.Prefix = m[:metricOffset]
	it := txn.NewIterator(o)
	defer it.Close()
	props := make(map[Property]*Filter)
	for _, e := range r.Filters {
		props[e.Property] = e
	}
	for p, f := range props {
		m.prop(p)
		out := make(MetricResult)
		for _, metric := range r.Metrics {
			if !f.Accept(metric) {
				continue
			}
			values := make(map[string]*internalValue)
			b.Reset()
			m.metric(metric)
			var text string
			if f.Expr.ExactMatch() {
				text = f.Expr.Text
			}
			// /user_id/site_id/metric/prop/text/
			prefix := m.idx(b, text).Bytes()
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				x := it.Item()
				if x.Version() > uint64(end) {
					// We have reached the end of iteration. Range actually
					// reflects on the version we are interested in.
					break
				}
				kb := x.Key()
				// text comes after the key offset
				txt := kb[keyOffset : len(kb)-6]
				if !f.Expr.Match(txt) {
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
			if len(values) == 0 {
				// No need to include empty metrics
				continue
			}
			o := make(OutValue)
			for k, v := range values {
				o[k] = rollUp(v.Value, v.Timestamp, shared, func(u []uint16) uint32 {
					var x uint32
					for _, n := range u {
						x += uint32(n)
					}
					return x
				})
			}
			out[metric.String()] = o
		}
		if len(out) == 0 {
			continue
		}
		result.Result[p.String()] = out
	}
	return
}

func Query2(ctx context.Context, uid, sid uint64, r query.Query) (result query.QueryResult) {
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
		o[k] = rollUp(v.Value, v.Timestamp, shared, func(u []uint16) uint32 {
			var x uint32
			for _, n := range u {
				x += uint32(n)
			}
			return x
		})
	}
}
