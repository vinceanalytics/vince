package ro2

import (
	"regexp"
	"strconv"
	"strings"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/roaring"
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
					Op:     roaring.EQ,
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
					Op:     roaring.EQ,
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
					Op:     roaring.EQ,
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
					Op:     roaring.EQ,
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

func (Reject) Apply(rtx *Tx, shard uint64, view uint64, columns *roaring.Bitmap) *roaring.Bitmap {
	return nil
}

type Filter interface {
	Apply(rtx *Tx, shard uint64, view uint64, columns *roaring.Bitmap) *roaring.Bitmap
}

type And []Filter

func (a And) Apply(rtx *Tx, shard uint64, view uint64, columns *roaring.Bitmap) *roaring.Bitmap {
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
	Op     roaring.Operation
}

func (m *Match) Apply(rtx *Tx, shard uint64, view uint64, columns *roaring.Bitmap) (b *roaring.Bitmap) {
	rtx.Bitmap(shard, view, m.Field, func(bs *roaring.BSI) {
		b = m.apply(bs, columns)
		if m.Negate {
			b = bs.GetExistenceBitmap().Clone()
			b.AndNot(m.apply(bs, columns))
		} else {
			b = m.apply(bs, columns)
		}
	})
	return b
}

func (m *Match) apply(bs *roaring.BSI, columns *roaring.Bitmap) *roaring.Bitmap {
	if len(m.Values) == 1 {
		m := bs.CompareValue(0, m.Op, m.Values[0], 0, bs.GetExistenceBitmap())
		return roaring.And(m, columns)
	}
	o := make([]*roaring.Bitmap, len(m.Values))

	for i := range m.Values {
		o[i] = roaring.And(
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
