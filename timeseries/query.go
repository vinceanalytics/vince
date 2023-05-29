package timeseries

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"regexp"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/gernest/vince/pkg/log"
	"github.com/gernest/vince/pkg/timex"
	"github.com/gernest/vince/system"
)

type QueryRequest struct {
	UserID   uint64
	SiteID   uint64
	Range    timex.Range
	Metrics  []Metric
	Property map[Property]*Match
}

type QueryResult struct {
	ELapsed time.Duration
	Result  PropResult
}

type Match struct {
	Text string
	IsRe bool
	re   *regexp.Regexp
}

// AggregateResult is a map of AggregateType to the value it represent.
type AggregateResult map[Metric][]Value

type Value struct {
	Timestamp uint64
	Value     uint32
}

type PropResult map[Property]PropValues

type PropValues map[string]AggregateResult

func (a AggregateResult) MarshalJSON() ([]byte, error) {
	o := make(map[string][]Value)
	for k, v := range a {
		o[k.String()] = v
	}
	return json.Marshal(o)
}

func (a PropResult) MarshalJSON() ([]byte, error) {
	o := make(map[string]PropValues)
	for k, v := range a {
		o[k.String()] = v
	}
	return json.Marshal(o)
}

func Query(ctx context.Context, r QueryRequest) (result QueryResult) {
	start := time.Now()
	txn := GetMike(ctx).NewTransaction(false)
	var startTS, endTS uint64
	if !r.Range.From.IsZero() {
		startTS = uint64(r.Range.From.Truncate(time.Hour).Unix())
	}
	if !r.Range.To.IsZero() {
		endTS = uint64(r.Range.To.Truncate(time.Hour).Unix())
	}
	m := newMetaKey()
	defer func() {
		m.Release()
		txn.Discard()
		result.ELapsed = time.Since(start)
		system.QueryDuration.Observe(result.ELapsed.Seconds())
	}()
	m.uid(r.UserID, r.SiteID)
	if len(r.Property) == 0 || len(r.Metrics) == 0 {
		return
	}
	b := smallBufferpool.Get().(*bytes.Buffer)
	defer func() {
		b.Reset()
		smallBufferpool.Put(b)
	}()
	result.Result = make(PropResult)
	o := badger.DefaultIteratorOptions
	if !r.Range.From.IsZero() {
		o.SinceTs = startTS
	}
	// make sure all iterations are in /user_id/site_id/ scope
	o.Prefix = m[:metricOffset]
	it := txn.NewIterator(o)
	defer it.Close()

	for p, match := range r.Property {
		values := make(PropValues)
		for _, metric := range r.Metrics {
			b.Reset()
			m.prop(p).metric(metric)
			if !match.IsRe {
				var text string
				if !match.IsRe {
					text = match.Text
				}
				// /user_id/site_id/metric/prop/text/
				prefix := m.idx(b, text).Bytes()
				for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
					x := it.Item()
					if endTS != 0 && x.Version() > endTS {
						// We have reached the end of iteration. Range actually
						// reflects on the version we are interested in.
						break
					}
					kb := x.Key()

					// last 6 bytes of the key are for the timestamp
					ts := kb[len(kb)-6:]
					// text comes before the timestamp
					txt := kb[keyOffset : len(kb)-6]
					if match.IsRe && !match.re.Match(txt) {
						continue
					}
					xv, ok := values[string(txt)]
					if !ok {
						xv = make(AggregateResult)
						values[string(txt)] = xv
					}
					err := x.Value(func(val []byte) error {
						xv[metric] = append(xv[metric], Value{
							Timestamp: Time(ts),
							Value:     binary.BigEndian.Uint32(val),
						})
						return nil
					})
					if err != nil {
						log.Get(ctx).Err(err).Msg("failed to read value from kv store")
					}
				}
			}
		}
		result.Result[p] = values
	}
	return
}
