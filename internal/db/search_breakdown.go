package db

import (
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/gernest/rbf/dsl"
	"github.com/gernest/rows"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
)

type breakdownQuery struct {
	props      map[v1.Property]map[uint64]*aggregate
	properties []v1.Property
	metrics    []v1.Metric
}

var _ Query = (*breakdownQuery)(nil)

func newBreakdown(props []v1.Property, m []v1.Metric) *breakdownQuery {
	return &breakdownQuery{
		props:   make(map[v1.Property]map[uint64]*aggregate),
		metrics: m,
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
