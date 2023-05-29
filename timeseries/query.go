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
	idx := GetIndex(ctx).NewTransaction(false)
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
		idx.Discard()
		result.ELapsed = time.Since(start)
		system.QueryDuration.Observe(result.ELapsed.Seconds())
	}()
	m.uid(r.UserID)
	m.sid(r.UserID)
	if len(r.Property) == 0 || len(r.Metrics) == 0 {
		return
	}
	b := smallBufferpool.Get().(*bytes.Buffer)
	key := smallBufferpool.Get().(*bytes.Buffer)
	defer func() {
		b.Reset()
		key.Reset()
		smallBufferpool.Put(b)
		smallBufferpool.Put(key)
	}()
	result.Result = make(PropResult)
	o := badger.IteratorOptions{}
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
				prefix := m.IndexBufferPrefix(b, text).Bytes()
				for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
					x := it.Item()
					if endTS != 0 && x.Version() > endTS {
						// We have reached the end of iteration
						break
					}
					key.Reset()
					mike, txt, ts := IndexToKey(x.Key(), key)
					if match.IsRe && !match.re.Match(txt) {
						continue
					}
					xv, ok := values[string(txt)]
					if !ok {
						xv = make(AggregateResult)
						values[string(txt)] = xv
					}
					mv, err := txn.Get(mike.Bytes())
					if err != nil {
						log.Get(ctx).Err(err).Msg("failed to get mike value")
						continue
					}
					mv.Value(func(val []byte) error {
						xv[metric] = append(xv[metric], Value{
							Timestamp: Time(ts),
							Value:     binary.BigEndian.Uint32(val),
						})
						return nil
					})
				}
			}
		}
		result.Result[p] = values
	}
	return
}
