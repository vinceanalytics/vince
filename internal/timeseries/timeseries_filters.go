package timeseries

import (
	"context"
	"encoding/binary"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/roaring"
	"github.com/vinceanalytics/vince/internal/util/data"
	wq "github.com/vinceanalytics/vince/internal/web/query"
)

func (ts *Timeseries) compile(fs wq.Filters) Filter {
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
					values[i] = int64(ts.Translate(fd, []byte(f.Value[i])))
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
						values = append(values, int64(ts.Translate(fd, []byte(source))))
					} else {
						re, err := regexp.Compile(source)
						if err != nil {
							return Reject{}
						}

						ts.Search(fd, prefix, func(key []byte, val uint64) {
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
				ts.Search(fd, []byte{}, func(b []byte, val uint64) {
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

func (Reject) Apply(ctx context.Context, rtx *Timeseries, shard uint64, view uint64, columns *roaring.Bitmap) *roaring.Bitmap {
	return nil
}

type Filter interface {
	Apply(ctx context.Context, rtx *Timeseries, shard uint64, view uint64, columns *roaring.Bitmap) *roaring.Bitmap
}

type And []Filter

func (a And) Apply(ctx context.Context, rtx *Timeseries, shard uint64, view uint64, columns *roaring.Bitmap) *roaring.Bitmap {
	if len(a) == 0 {
		return columns
	}
	if len(a) == 1 {
		return a[0].Apply(ctx, rtx, shard, view, columns)
	}
	m := a[0].Apply(ctx, rtx, shard, view, columns)
	for _, h := range a[1:] {
		m.And(h.Apply(ctx, rtx, shard, view, columns))
	}
	return m
}

type Match struct {
	Values []int64
	Field  models.Field
	Negate bool
}

func (m *Match) Apply(ctx context.Context, rtx *Timeseries, shard uint64, view uint64, columns *roaring.Bitmap) (b *roaring.Bitmap) {
	if m.Negate {
		return roaring.NewBitmap()
	}
	bs := rtx.NewBitmap(ctx, shard, view, m.Field)
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

func (ts *Timeseries) Search(field models.Field, prefix []byte, f func(key []byte, value uint64)) {
	sk := encoding.TranslateKey(field, prefix)
	data.Prefix(ts.db, sk, func(key, value []byte) error {
		f(key[3:], binary.BigEndian.Uint64(value))
		return nil
	})
}

func searchPrefix(source []byte) (prefix []byte, exact bool) {
	for i := range source {
		if special(source[i]) {
			return source[:i], false
		}
	}
	return source, true
}

// Bitmap used by func special to check whether a character needs to be escaped.
var specialBytes [16]byte

// special reports whether byte b needs to be escaped by QuoteMeta.
func special(b byte) bool {
	return b < utf8.RuneSelf && specialBytes[b%16]&(1<<(b/16)) != 0
}

func init() {
	for _, b := range []byte(`\.+*?()|[]{}^$`) {
		specialBytes[b%16] |= 1 << (b / 16)
	}
}
