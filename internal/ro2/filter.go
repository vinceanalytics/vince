package ro2

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/vinceanalytics/vince/internal/bsi"
	"github.com/vinceanalytics/vince/internal/models"
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
					Field:  models.Field_city,
					Op:     bsi.EQ,
					Values: []int64{int64(code)},
				}
			}
		default:
			fd := models.Field(models.Field_value[f.Key])
			if fd == 0 {
				return Reject{}
			}

			switch f.Op {
			case "is", "is_not":
				values := make([]int64, len(f.Value))
				for i := range f.Value {
					values[i] = int64(tx.tr(fd, []byte(f.Value[i])))
				}
				a = append(a, &Match{
					Field:  fd,
					Negate: f.Op == "is_not",
					Op:     bsi.EQ,
					Values: values,
				})
			case "matches", "does_not_match":
				var values []int64
				for _, source := range f.Value {
					prefix, exact := searchPrefix([]byte(source))
					if exact {
						values = append(values, int64(tx.tr(fd, []byte(source))))
					} else {
						re, err := regexp.Compile(source)
						if err != nil {
							return Reject{}
						}
						tx.Search(fd, prefix, func(b []byte, val uint64) {
							if re.Match(b) {
								values = append(values, int64(val))
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
					Op:     bsi.EQ,
					Values: values,
				})
			case "contains", "does_not_contain":
				var values []int64
				re, err := regexp.Compile(strings.Join(f.Value, "|"))
				if err != nil {
					return Reject{}
				}
				tx.Search(fd, []byte{}, func(b []byte, val uint64) {
					if re.Match(b) {
						values = append(values, int64(val))
					}
				})
				a = append(a, &Match{
					Field:  fd,
					Negate: f.Op == "does_not_contain",
					Op:     bsi.EQ,
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
	Values []int64
	Op     bsi.Operation
	Field  models.Field
	Negate bool
}

func (m *Match) Apply(rtx *Tx, shard uint64, view uint64, columns *roaring.Bitmap) (b *roaring.Bitmap) {
	bs := rtx.Bitmap(shard, view, m.Field)
	if m.Negate {
		b = bs.GetExistenceBitmap().Clone()
		b.AndNot(m.apply(bs, columns))
		return b
	}
	return m.apply(bs, columns)
}

func (m *Match) apply(bs *bsi.BSI, columns *roaring.Bitmap) *roaring.Bitmap {
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
