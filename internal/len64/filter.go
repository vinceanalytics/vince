package len64

import (
	"hash/crc32"

	"github.com/RoaringBitmap/roaring/v2/roaring64"
)

type Filter interface {
	Apply(db *View, foundSet *roaring64.Bitmap) (*roaring64.Bitmap, error)
}

type OP uint

const (
	EQ OP = iota
	NEQ
	RE
	NRE
)

type NoopFilter struct{}

func (NoopFilter) Apply(db *View, foundSet *roaring64.Bitmap) (*roaring64.Bitmap, error) {
	return foundSet, nil
}

type Text struct {
	OP    OP
	Field string
	Value string
}

func (s *Text) Apply(db *View, foundSet *roaring64.Bitmap) (*roaring64.Bitmap, error) {
	b, err := db.Get(s.Field)
	if err != nil {
		return nil, err
	}
	switch s.OP {
	case EQ, NEQ:
		h := crc32.NewIEEE()
		h.Write([]byte(s.Field))
		h.Write(sep)
		h.Write([]byte(s.Value))
		value := h.Sum32()
		r := b.CompareValue(parallel(), roaring64.EQ, int64(value), 0, foundSet)
		if s.OP == NEQ {
			return roaring64.AndNot(r, foundSet), nil
		}
		return r, nil
	case RE, NRE:
		values, err := SearchRegex(db.snap, db.shard, s.Field, s.Value)
		if err != nil {
			return nil, err
		}
		r := roaring64.New()
		it := values.Iterator()

		for it.HasNext() {
			m := int64(it.Next())
			o := b.CompareValue(parallel(), roaring64.EQ, m, 0, foundSet)
			r.Or(o)
		}
		if s.OP == NRE {
			return roaring64.AndNot(r, foundSet), nil
		}
		return r, nil
	default:
		return roaring64.New(), nil
	}
}
