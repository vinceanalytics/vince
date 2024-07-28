package len64

import "github.com/RoaringBitmap/roaring/v2/roaring64"

type Projection map[string]*roaring64.BSI

func (p Projection) Visitors() uint64 {
	uniq := p["uid"].Transpose()
	return uniq.GetCardinality()
}

func (p Projection) Visits() uint64 {
	b := p["session"]
	sum, _ := b.Sum(b.GetExistenceBitmap())
	return uint64(sum)
}

func (p Projection) Bounce() uint64 {
	b := p["bounce"]
	sum, _ := b.Sum(b.GetExistenceBitmap())
	return uint64(sum)
}

type Group struct {
	Key        string
	Value      int64
	Projection Projection
}

func (p Projection) GroupBy(name string) []Group {
	b := p[name]
	uniq := b.Transpose()
	it := uniq.Iterator()
	o := make([]Group, 0, uniq.GetCardinality())
	for it.HasNext() {
		value := int64(it.Next())
		r := b.CompareValue(parallel(), roaring64.EQ, value, value, b.GetExistenceBitmap())
		o = append(o, Group{
			Key:        name,
			Value:      value,
			Projection: p.clone(r, name),
		})
	}
	return o
}

func (p Projection) clone(foundSet *roaring64.Bitmap, skip string) Projection {
	o := make(Projection)
	for k, v := range p {
		if _, ok := p[skip]; ok {
			continue
		}
		b := v.NewBSIRetainSet(foundSet)
		o[k] = b
	}
	return o
}
