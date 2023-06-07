package timeseries

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"path"
	"regexp"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/vinceanalytics/vince/internal/core"
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

var (
	defaultStep = time.Minute * 5
)

type PropertiesResult map[string]MetricResult

type MetricResult map[string]OutValue

type OutValue map[string][]float64

func Query(ctx context.Context, r QueryRequest) (result QueryResult) {
	currentTime := core.Now(ctx)
	startTS := currentTime.Truncate(time.Millisecond)
	if !r.Start.IsZero() {
		startTS = r.Start.Truncate(time.Millisecond)
	}

	step := defaultStep
	if r.Step > 0 {
		step = r.Step
	}

	var window, offset time.Duration
	if r.Window > 0 {
		window = r.Window
	}
	if window == 0 {
		// default to session ttl
		window = time.Minute * 30
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
	windowMs := window.Milliseconds()

	shared := sharedTS(start, end, step.Milliseconds())
	result.Timestamps = shared
	result.Result = make(PropertiesResult)

	txn := GetMike(ctx).NewTransactionAt(uint64(start), false)

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
			values := make(map[string]*Value)
			for _, metric := range r.Metrics {
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
					txt := kb[keyOffset:]
					if !f.Expr.Match(txt) {
						continue
					}
					v, ok := values[string(text)]
					if !ok {
						v = &Value{}
						values[string(text)] = v
					}
					err := x.Value(func(val []byte) error {
						v.Timestamp = append(v.Timestamp, int64(x.Version()))
						v.Value = append(v.Value,
							math.Float64frombits(binary.BigEndian.Uint64(val)),
						)
						return nil
					})
					if err != nil {
						log.Get().Err(err).Msg("failed to read value from kv store")
					}
				}

			}
			o := make(OutValue)
			for k, v := range values {
				o[k] = rollUp(windowMs, v.Value, v.Timestamp, shared, func(ro *rollOptions) float64 {
					var x float64
					for _, xx := range ro.values {
						x += xx
					}
					return x
				})
			}
			out[metric.String()] = o
		}
		result.Result[p.String()] = out
	}
	return
}
