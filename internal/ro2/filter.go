package ro2

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/RoaringBitmap/roaring/v2/roaring64"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	wq "github.com/vinceanalytics/vince/internal/web/query"
)

func (tx *Tx) compile(fs wq.Filters) Filter {
	a := make(And, 0, len(fs))
	for _, f := range fs {
		switch f.Key {
		case "city":
			switch f.Op {
			case "is":
				code, _ := strconv.Atoi(f.Value[0])
				if code == 0 {
					return Reject{}
				}
				return &Match{
					Field:  v1.Field_city,
					Op:     roaring64.EQ,
					Values: []int64{int64(code)},
				}
			}
		default:
			fd := v1.Field(v1.Field_value[f.Key])
			if fd == 0 {
				return Reject{}
			}

			switch f.Op {
			case "is", "is_not":
				values := tx.ids(fd, f.Value)
				if len(values) == 0 {
					return Reject{}
				}
				a = append(a, &Match{
					Field:  fd,
					Negate: f.Op == "is_not",
					Op:     roaring64.EQ,
					Values: values,
				})
			case "matches", "does_not_match":
				var values []int64
				for _, source := range f.Value {
					prefix, exact := searchPrefix([]byte(source))
					if exact {
						id, ok := tx.ID(fd, source)
						if ok {
							values = append(values, int64(id))
						}
					} else {
						re, err := regexp.Compile(source)
						if err != nil {
							return Reject{}
						}
						tx.Search(fd, prefix, func(b []byte, u uint64) {
							if re.Match(b) {
								values = append(values, int64(u))
							}
						})
					}
				}
				if len(values) == 0 {
					return Reject{}
				}
				a = append(a, &Match{
					Field:  fd,
					Negate: f.Op == "does_not_match",
					Op:     roaring64.EQ,
					Values: values,
				})
			case "contains", "does_not_contain":
				var values []int64
				re, err := regexp.Compile(strings.Join(f.Value, "|"))
				if err != nil {
					return Reject{}
				}
				tx.Search(fd, []byte{}, func(b []byte, u uint64) {
					if re.Match(b) {
						values = append(values, int64(u))
					}
				})
				a = append(a, &Match{
					Field:  fd,
					Negate: f.Op == "does_not_contain",
					Op:     roaring64.EQ,
					Values: values,
				})
			default:
				return Reject{}
			}
		}
	}
	return a
}

type Reject struct{}

func (Reject) Apply(rtx *Tx, shard uint64, view uint64, columns *roaring64.Bitmap) *roaring64.Bitmap {
	return roaring64.New()
}

type Filter interface {
	Apply(rtx *Tx, shard uint64, view uint64, columns *roaring64.Bitmap) *roaring64.Bitmap
}

type And []Filter

func (a And) Apply(rtx *Tx, shard uint64, view uint64, columns *roaring64.Bitmap) *roaring64.Bitmap {
	if len(a) == 0 {
		return columns
	}
	if len(a) == 1 {
		return a[0].Apply(rtx, shard, view, columns)
	}
	m := a[0].Apply(rtx, shard, view, columns)
	for _, h := range a[1:] {
		m.And(h.Apply(rtx, shard, view, columns))
	}
	return m
}

type Match struct {
	Field  v1.Field
	Values []int64
	Negate bool
	Op     roaring64.Operation
}

func (m *Match) Apply(rtx *Tx, shard uint64, view uint64, columns *roaring64.Bitmap) *roaring64.Bitmap {
	bs, err := rtx.Bitmap(shard, view, m.Field)
	if err != nil {
		return roaring64.New()
	}
	b := m.apply(bs, columns)
	if m.Negate {
		return roaring64.AndNot(bs.GetExistenceBitmap(), b)
	}
	return b
}

func (m *Match) apply(bs *roaring64.BSI, columns *roaring64.Bitmap) *roaring64.Bitmap {
	if len(m.Values) == 1 {
		m := bs.CompareValue(0, m.Op, m.Values[0], 0, bs.GetExistenceBitmap())
		return roaring64.And(m, columns)
	}
	o := make([]*roaring64.Bitmap, len(m.Values))

	for i := range m.Values {
		o[i] = roaring64.And(
			bs.CompareValue(0, m.Op, m.Values[0], 0, bs.GetExistenceBitmap()),
			columns,
		)
	}
	b := o[0]
	for _, n := range o[1:] {
		b.Or(n)
	}
	return b
}
