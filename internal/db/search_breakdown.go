package db

import (
	"cmp"
	"context"
	"slices"
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/gernest/rbf/dsl"
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
	query := newBreakdown(req.Property, req.Metrics)
	from, to := periodToRange(req.Period, req.Date)
	err = db.Search(from, to, append(req.Filters, &v1.Filter{
		Property: v1.Property_domain,
		Op:       v1.Filter_equal,
		Value:    req.SiteId,
	}), query)
	if err != nil {
		return nil, err
	}
	return &v1.BreakDown_Response{Results: query.result}, nil
}

type breakdownQuery struct {
	props      map[v1.Property]map[uint64]*aggregate
	properties []v1.Property
	metrics    []v1.Metric
	result     []*v1.BreakDown_Result
}

func (b *breakdownQuery) Final(tx *Tx) error {
	b.result = make([]*v1.BreakDown_Result, 0, len(b.props))
	for _, prop := range b.properties {
		r := &v1.BreakDown_Result{Property: prop}
		p := b.props[prop]
		r.Values = make([]*v1.BreakDown_KeyValues, 0, len(p))
		for k, v := range p {
			x := &v1.BreakDown_KeyValues{
				Key:   tx.Tr(prop.String(), k),
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
	return nil
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

func (b *breakdownQuery) Apply(tx *Tx, columns *rows.Row) error {
	f := new(ViewFmt)
	o := roaring64.New()
	add := func(_, value uint64) error {
		o.Add(value)
		return nil
	}

	for _, prop := range b.properties {
		view := f.Format(tx.View, prop.String())
		// find all unique properties
		o.Clear()
		err := dsl.ExtractValuesBSI(tx.Tx, view, tx.Shard, columns, add)
		if err != nil {
			return err
		}
		if o.IsEmpty() {
			continue
		}
		it := o.Iterator()
		for it.HasNext() {
			id := it.Next()

			// Find all columns for this id
			r, err := dsl.CompareValueBSI(tx.Tx, view, tx.Shard, dsl.EQ, id, 0, columns)
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
		a = newAggregate()
		a.View(b.metrics)
		m[id] = a
	}
	return a
}
