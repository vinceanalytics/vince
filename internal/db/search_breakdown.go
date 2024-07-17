package db

import (
	"cmp"
	"context"
	"slices"

	"github.com/gernest/roaring"
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
	err = db.view(from, to, req.TenantId, func(tx *view, r *rows.Row) error {
		r, err := tx.filters(req.Filters, r)
		if err != nil {
			return err
		}
		if r.IsEmpty() {
			return nil
		}
		return a.Apply(tx, tx.shard, r)
	}, a.Final)
	if err != nil {
		return nil, err
	}
	return &v1.BreakDown_Response{Results: a.result}, nil
}

type breakdownQuery struct {
	props      map[v1.Property]map[uint64]*aggregate
	properties []v1.Property
	metrics    []v1.Metric
	result     []*v1.BreakDown_Result
}

func (b *breakdownQuery) Final(tx *view) error {
	b.result = make([]*v1.BreakDown_Result, 0, len(b.props))
	for _, prop := range b.properties {
		r := &v1.BreakDown_Result{Property: prop}
		p := b.props[prop]
		r.Values = make([]*v1.BreakDown_KeyValues, 0, len(p))
		for k, v := range p {
			x := &v1.BreakDown_KeyValues{
				Key:   string(tx.key(prop.String(), k)),
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

func newBreakdown(props []v1.Property, m []v1.Metric) *breakdownQuery {
	return &breakdownQuery{
		props:      make(map[v1.Property]map[uint64]*aggregate),
		metrics:    dupe(m),
		properties: dupe(props),
	}
}

func (b *breakdownQuery) Apply(tx *view, shard uint64, filters *rows.Row) error {
	filter := make([]*roaring.Container, 1<<shardVsContainerExponent)
	filterIterator, _ := filters.Segments[0].Data().Containers.Iterator(0)
	// So let's get these all with a nice convenient 0 offset...
	for filterIterator.Next() {
		k, c := filterIterator.Value()
		if c.N() == 0 {
			continue
		}
		filter[k%(1<<shardVsContainerExponent)] = c
	}
	for _, prop := range b.properties {
		rs, err := tx.distinct(prop.String(), filter)
		if err != nil {
			return err
		}
		if len(rs) == 0 {
			continue
		}
		for id, r := range rs {
			// Compute aggregates for columns belonging to id
			err = b.get(prop, id).cache.Apply(tx, shard, rows.NewRow(r...))
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
