package timeseries

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"math"
	"regexp"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/gernest/vince/pkg/log"
	"github.com/gernest/vince/pkg/timex"
)

type QueryRequest struct {
	UserID   uint64
	SiteID   uint64
	Range    timex.Range
	Metrics  []Metric
	Property map[Property]*Match
}

type Match struct {
	Text string
	IsRe bool
	re   *regexp.Regexp
}

// AggregateResult is a map of AggregateType to the value it represent.
type AggregateResult map[Metric][]Value

type Value struct {
	Timestamp int64
	Value     float64
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

func Query(ctx context.Context, r QueryRequest) (result PropResult) {
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
		defer m.Release()
		txn.Discard()
		idx.Discard()
	}()
	m.SetUserID(r.UserID)
	m.SetSiteID(r.UserID)
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
	result = make(PropResult)
	for p, match := range r.Property {
		values := make(PropValues)
		for _, metric := range r.Metrics {
			b.Reset()
			m.Prop(p).Metric(metric)
			if !match.IsRe {
				var text string
				if !match.IsRe {
					text = match.Text
				}
				prefix := m.IndexBufferPrefix(b, text).Bytes()
				o := badger.IteratorOptions{}
				o.Prefix = prefix
				if !r.Range.From.IsZero() {
					o.SinceTs = startTS
				}
				it := txn.NewIterator(o)
				for it.Rewind(); it.Valid(); it.Next() {
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
							Timestamp: timex.FromTimestamp(ts),
							Value: math.Float64frombits(
								binary.BigEndian.Uint64(val),
							),
						})
						return nil
					})
				}
				it.Close()
			}
		}
		result[p] = values
	}
	return
}
