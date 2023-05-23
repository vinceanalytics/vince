package timeseries

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"math"
	"regexp"

	"github.com/dgraph-io/badger/v4"
	"github.com/gernest/vince/pkg/log"
	"github.com/gernest/vince/pkg/timex"
)

func (a METRIC_TYPE) MarshalJSON() ([]byte, error) {
	return []byte(a.String()), nil
}

func (a PROPS) MarshalJSON() ([]byte, error) {
	return []byte(a.String()), nil
}

type QueryRequest struct {
	UserID  uint64
	SiteID  uint64
	Range   timex.Range
	Metrics []METRIC_TYPE
	Props   map[PROPS]*Match
}

type Match struct {
	Text string
	IsRe bool
	re   *regexp.Regexp
}

func (m *Match) Match(o []byte) bool {
	if m.IsRe {
		return m.re.Match(o)
	}
	return m.Text == string(o)
}

// AggregateResult is a map of AggregateType to the value it represent.
type AggregateResult map[METRIC_TYPE]Value

type Value struct {
	Timestamp int64
	Value     float64
}

type PropResult map[PROPS]PropValues

type PropValues map[string]AggregateResult

func (a AggregateResult) MarshalJSON() ([]byte, error) {
	o := make(map[string]Value)
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

func Query(ctx context.Context, request QueryRequest) (r PropResult) {
	db := GetMike(ctx)
	m := newMetaKey()
	defer func() {
		defer m.Release()
	}()
	m.SetUserID(request.UserID)
	m.SetSiteID(request.UserID)
	if len(request.Props) == 0 {
		return
	}
	for k, v := range request.Props {
		m.SetProp(k)
		// Passing this means we also include root stats
		err := db.View(func(txn *badger.Txn) error {
			o := badger.DefaultIteratorOptions
			o.PrefetchValues = false
			if !request.Range.From.IsZero() {
				o.SinceTs = uint64(request.Range.From.Unix())
			}
			agg := make(PropValues)
			b := smallBufferpool.Get().(*bytes.Buffer)
			for _, mt := range request.Metrics {
				b.Reset()
				n := o
				m.SetAggregateType(mt)
				if !v.IsRe {
					// we are doing exact match
					key := m.StringBuffer(b, v.Text).Bytes()
					x, err := txn.Get(key)
					if err != nil {
						if errors.Is(err, badger.ErrKeyNotFound) {
							continue
						}
						return err
					}
					x.Value(func(val []byte) error {
						ks := x.Key()[hashOffset:]
						value := math.Float64frombits(
							binary.BigEndian.Uint64(val),
						)
						if mx, ok := agg[string(ks)]; ok {
							mx[mt] = Value{
								Timestamp: timex.FromTimestamp(key[yearOffset:]),
								Value:     value,
							}
						}
						return nil
					})
				} else {
					n.Prefix = m[:yearOffset]
					it := txn.NewIterator(o)
					for it.Rewind(); it.Valid(); it.Next() {
						x := it.Item()
						key := x.Key()
						if v.Match(key[hashOffset:]) {
							x.Value(func(val []byte) error {
								xk := x.Key()[hashOffset:]
								value := math.Float64frombits(
									binary.BigEndian.Uint64(val),
								)
								ts := timex.FromTimestamp(key[yearOffset:])
								if mx, ok := agg[string(xk)]; ok {
									mx[mt] = Value{
										Timestamp: ts,
										Value:     value,
									}
								} else {
									agg[string(xk)] = AggregateResult{
										mt: Value{
											Timestamp: ts,
											Value:     value,
										},
									}
								}
								return nil
							})
						}
					}
					it.Close()
				}
			}
			if r == nil {
				r = make(PropResult)
				r[k] = agg
			}
			return nil
		})
		if err != nil {
			log.Get(ctx).Err(err).Msg("failed to query")
		}
	}
	return
}
