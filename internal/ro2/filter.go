package ro2

import (
	"regexp"
	"strconv"
	"strings"

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
					values[i] = int64(tx.store.ID(fd, []byte(f.Value[i])))
				}
				a = append(a, &Match{
					Field:  fd,
					Negate: f.Op == "is_not",
					Values: values,
				})
			case "matches", "does_not_match":
				var values []int64
				for _, source := range f.Value {
					prefix, exact := searchPrefix([]byte(source))
					if exact {
						values = append(values, int64(tx.store.ID(fd, []byte(source))))
					} else {
						re, err := regexp.Compile(source)
						if err != nil {
							return Reject{}
						}

						tx.store.Search(fd, prefix, func(key []byte, val uint64) {
							if re.Match(key) {
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
					Values: values,
				})
			case "contains", "does_not_contain":
				var values []int64
				re, err := regexp.Compile(strings.Join(f.Value, "|"))
				if err != nil {
					return Reject{}
				}
				tx.store.Search(fd, []byte{}, func(b []byte, val uint64) {
					if re.Match(b) {
						values = append(values, int64(val))
					}
				})
				a = append(a, &Match{
					Field:  fd,
					Negate: f.Op == "does_not_contain",
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
	Field  models.Field
	Negate bool
}

func (m *Match) Apply(rtx *Tx, shard uint64, view uint64, columns *roaring.Bitmap) (b *roaring.Bitmap) {
	if m.Negate {
		return roaring.NewBitmap()
	}
	bs := rtx.NewBitmap(shard, view, m.Field)
	return m.apply(bs, shard, columns)
}

func (m *Match) apply(bs *roaring.Bitmap, shard uint64, columns *roaring.Bitmap) *roaring.Bitmap {
	if len(m.Values) == 1 {
		row := bs.Row(shard, uint64(m.Values[0]))
		row.And(columns)
		return row
	}
	o := make([]*roaring.Bitmap, len(m.Values))

	for i := range m.Values {
		row := bs.Row(shard, uint64(m.Values[i]))
		row.And(columns)
		o[i] = row
	}
	return roaring.FastOr(o...)
}
