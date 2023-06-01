package timeseries

import (
	"bytes"
	"context"
	"encoding/binary"
	"regexp"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/gernest/vince/pkg/log"
	"github.com/gernest/vince/pkg/timex"
	"github.com/gernest/vince/system"
)

type BaseQuery struct {
	Range    timex.Range `json:"range"`
	Metrics  []Metric    `json:"metrics"`
	Property Property    `json:"prop"`
	Match    Match       `json:"match"`
}

type QueryRequest struct {
	UserID uint64
	SiteID uint64
	BaseQuery
}

type QueryResult struct {
	ELapsed time.Duration `json:"elapsed"`
	Result  []Result      `json:"result"`
}

type Match struct {
	Text string         `json:"text"`
	IsRe bool           `json:"isRe"`
	Re   *regexp.Regexp `json:"-"`
}

type Value struct {
	Timestamp uint64 `json:"timestamp"`
	Value     uint32 `json:"value"`
}

type Result struct {
	Metric Metric             `json:"metric"`
	Values map[string][]Value `json:"values"`
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
	if !r.Range.From.IsZero() {
		o.SinceTs = startTS
	}
	// make sure all iterations are in /user_id/site_id/ scope
	o.Prefix = m[:metricOffset]
	it := txn.NewIterator(o)
	defer it.Close()
	for _, metric := range r.Metrics {
		values := make(map[string][]Value)
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
				if r.Match.IsRe && !r.Match.Re.Match(txt) {
					continue
				}
				err := x.Value(func(val []byte) error {
					values[string(txt)] = append(values[string(txt)], Value{
						Timestamp: Time(ts),
						Value:     binary.BigEndian.Uint32(val),
					})
					return nil
				})
				if err != nil {
					log.Get().Err(err).Msg("failed to read value from kv store")
				}
			}

		}
		result.Result = append(result.Result, Result{
			Metric: metric,
			Values: values,
		})
	}
	return
}
