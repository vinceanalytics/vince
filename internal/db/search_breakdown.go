package db

import (
	"cmp"
	"context"
	"slices"

	"github.com/RoaringBitmap/roaring/roaring64"
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
	o := roaring64.New()
	var filterBitmap *roaring.Bitmap
	if filters != nil && len(filters.Segments) > 0 {
		filterBitmap = filters.Segments[0].Data()
	}
	// We can't grab the containers "for each row" from the set-type field,
	// because we don't know how many rows there are, and some of them
	// might be empty, so really, we're going to iterate through the
	// containers, and then intersect them with the filter if present.
	var filter []*roaring.Container
	if filterBitmap != nil {
		filter = make([]*roaring.Container, 1<<shardVsContainerExponent)
		filterIterator, _ := filterBitmap.Containers.Iterator(0)
		// So let's get these all with a nice convenient 0 offset...
		for filterIterator.Next() {
			k, c := filterIterator.Value()
			if c.N() == 0 {
				continue
			}
			filter[k%(1<<shardVsContainerExponent)] = c
		}
	}

	for _, prop := range b.properties {
		o.Clear()
		err := tx.distinct(prop.String(), o, filter, filterBitmap != nil)
		if err != nil {
			return err
		}
		if o.IsEmpty() {
			continue
		}

		it := o.Iterator()

		for it.HasNext() {
			id := it.Next()
			r, err := tx.row(prop.String(), shard, id, filters)
			if err != nil {
				return err
			}
			// Compute aggregates for columns belonging to id
			err = b.get(prop, id).cache.Apply(tx, shard, r)
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
