package db

import (
	"cmp"
	"context"
	"slices"
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/gernest/rbf/dsl/bsi"
	"github.com/gernest/rbf/dsl/mutex"
	"github.com/gernest/rbf/dsl/tr"
	"github.com/gernest/rbf/dsl/tx"
	"github.com/gernest/rows"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/defaults"
)

func (db *DB) Breakdown(ctx context.Context, req *v1.BreakDown_Request) (*v1.BreakDown_Response, error) {
	defaults.Set(req)
	err := validate.Validate(req)
	if err != nil {
		return nil, err
	}
	a := newBreakdown(req.Property, req.Metrics)
	from, to := periodToRange(req.Period, req.Date)
	props := append(req.Filters,
		&v1.Filter{Property: v1.Property_domain, Op: v1.Filter_equal, Value: req.SiteId},
	)
	ts := bsi.Filter("timestamp", bsi.RANGE, from.UnixMilli(), to.UnixMilli())
	fs := filterProperties(props...)
	r, err := db.db.Reader()
	if err != nil {
		return nil, err
	}
	defer r.Release()
	for _, shard := range r.Range(from, to) {
		err := r.View(shard, func(txn *tx.Tx) error {
			f, err := ts.Apply(txn, nil)
			if err != nil {
				return err
			}
			if f.IsEmpty() {
				return nil
			}
			r, err := fs.Apply(txn, f)
			if err != nil {
				return err
			}
			if r.IsEmpty() {
				return nil
			}
			return a.Apply(txn, r)
		})
		if err != nil {
			return nil, err
		}
	}
	a.Final(r.Tr())

	return &v1.BreakDown_Response{Results: a.result}, nil
}

type breakdownQuery struct {
	props      map[v1.Property]map[uint64]*aggregate
	properties []v1.Property
	metrics    []v1.Metric
	result     []*v1.BreakDown_Result
}

func (b *breakdownQuery) Final(tr *tr.Read) {
	b.result = make([]*v1.BreakDown_Result, 0, len(b.props))
	for _, prop := range b.properties {
		r := &v1.BreakDown_Result{Property: prop}
		p := b.props[prop]
		r.Values = make([]*v1.BreakDown_KeyValues, 0, len(p))
		for k, v := range p {
			x := &v1.BreakDown_KeyValues{
				Key:   string(tr.Key(prop.String(), k)),
				Value: map[string]float64{},
			}
			for i := range b.metrics {
				x.Value[b.metrics[i].String()] = v.Result(b.metrics[i])
			}
			r.Values = append(r.Values, x)
		}
		slices.SortFunc(r.Values, func(a, b *v1.BreakDown_KeyValues) int {
			return cmp.Compare(a.Key, b.Key)
		})
		b.result = append(b.result, r)
	}
}

var _ Query = (*breakdownQuery)(nil)

func newBreakdown(props []v1.Property, m []v1.Metric) *breakdownQuery {
	return &breakdownQuery{
		props:      make(map[v1.Property]map[uint64]*aggregate),
		metrics:    dupe(m),
		properties: dupe(props),
	}
}

func (b *breakdownQuery) View(_ time.Time) View {
	return b
}

func (b *breakdownQuery) Apply(tx *tx.Tx, columns *rows.Row) error {
	o := roaring64.New()
	for _, prop := range b.properties {
		o.Clear()
		err := mutex.Distinct(tx, prop.String(), o, columns)
		if err != nil {
			return err
		}
		if o.IsEmpty() {
			continue
		}

		it := o.Iterator()

		for it.HasNext() {
			id := it.Next()
			r, err := mutex.Filter(prop.String(), id).Apply(tx, columns)
			if err != nil {
				return err
			}
			// Compute aggregates for columns belonging to id
			err = b.get(prop, id).cache.Apply(tx, r)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *breakdownQuery) get(prop v1.Property, id uint64) *aggregate {
	m, ok := b.props[prop]
	if !ok {
		m = make(map[uint64]*aggregate)
		b.props[prop] = m
	}
	a, ok := m[id]
	if !ok {
		a = newAggregate(b.metrics)
		m[id] = a
	}
	return a
}
