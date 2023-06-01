package timeseries

import (
	"bytes"
	"context"
	"encoding/binary"
	"math"
	"regexp"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/gernest/vince/pkg/log"
	"github.com/gernest/vince/system"
)

type BaseQuery struct {
	Start    time.Time     `json:"start,omitempty"`
	Offset   time.Duration `json:"offset,omitempty"`
	Step     time.Duration `json:"step,omitempty"`
	Window   time.Duration `json:"window,omitempty"`
	Metrics  []Metric      `json:"metrics"`
	Property Property      `json:"prop"`
	Match    Match         `json:"match"`
}

type QueryRequest struct {
	UserID uint64
	SiteID uint64
	BaseQuery
}

type QueryResult struct {
	ELapsed time.Duration `json:"elapsed"`
	Result  Output        `json:"result"`
}

type Match struct {
	Text string         `json:"text"`
	IsRe bool           `json:"isRe"`
	Re   *regexp.Regexp `json:"-"`
}

type Value struct {
	Timestamp []int64   `json:"timestamp"`
	Value     []float64 `json:"value"`
}

var (
	defaultStep = time.Minute * 5
)

type Output struct {
	Timestamps     []int64  `json:"timestamps"`
	Visitors       OutValue `json:"visitors,omitempty"`
	Views          OutValue `json:"views,omitempty"`
	Events         OutValue `json:"events,omitempty"`
	Visits         OutValue `json:"visits,omitempty"`
	BounceRates    OutValue `json:"bounceRates,omitempty"`
	VisitDurations OutValue `json:"visitDurations,omitempty"`
}

type OutValue map[string][]float64

func Query(ctx context.Context, r QueryRequest) (result QueryResult) {
	currentTime := time.Now()
	ct := currentTime.Truncate(time.Second).Unix()

	start := ct
	if !r.Start.IsZero() {
		start = r.Start.Truncate(time.Second).Unix()
	}
	step := int64(defaultStep)
	if r.Step > 0 {
		step = int64(r.Step)
	}

	var window, offset int64
	if r.Window > 0 {
		window = int64(r.Window)
	}
	if window == 0 {
		// default to session ttl
		window = int64(time.Minute * 30)
	}
	if r.Offset > 0 {
		offset = int64(r.Offset)
	}
	start -= offset
	end := start
	start = end - window
	start++
	if end < start {
		end = start
	}

	shared := sharedTS(start, end, step)
	result.Result.Timestamps = shared

	txn := GetMike(ctx).NewTransaction(false)

	m := newMetaKey()
	defer func() {
		m.Release()
		txn.Discard()
		result.ELapsed = time.Since(currentTime)
		system.QueryDuration.Observe(result.ELapsed.Seconds())
	}()
	m.uid(r.UserID, r.SiteID).prop(r.Property)
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
	for _, metric := range r.Metrics {
		values := make(map[string]*Value)
		for _, metric := range r.Metrics {
			b.Reset()
			m.metric(metric)
			var text string
			if !r.Match.IsRe {
				text = r.Match.Text
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
				if r.Match.IsRe && !r.Match.Re.Match(txt) {
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
			o[k] = rollUp(window, v.Value, v.Timestamp, shared, func(ro *rollOptions) float64 {
				var x float64
				for _, xx := range ro.values {
					x += xx
				}
				return x
			})
		}
		result.Result.Views = o
		switch metric {
		case Visitors:
			result.Result.Visitors = o
		case Views:
			result.Result.Views = o
		case Events:
			result.Result.Events = o
		case Visits:
			result.Result.Visits = o
		case BounceRates:
			result.Result.BounceRates = o
		case VisitDurations:
			result.Result.VisitDurations = o
		}
	}
	return
}
