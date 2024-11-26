package timeseries

import (
	"encoding/binary"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/ro2"
	"github.com/vinceanalytics/vince/internal/timeseries/cursor"
	"github.com/vinceanalytics/vince/internal/util/data"
	wq "github.com/vinceanalytics/vince/internal/web/query"
)

type Filter interface {
	Apply(cu *cursor.Cursor, re encoding.Resolution, shard, view uint64) *ro2.Bitmap
}

type And struct {
	Left, Right Filter
}

func (a And) Apply(cu *cursor.Cursor, re encoding.Resolution, shard, view uint64) *ro2.Bitmap {
	if a.Left == nil {
		return a.Right.Apply(cu, re, shard, view)
	}
	left := a.Left.Apply(cu, re, shard, view)
	if !left.Any() {
		return left
	}
	return left.Intersect(a.Right.Apply(cu, re, shard, view))
}

type Or struct {
	Left, Right Filter
}

func (o Or) Apply(cu *cursor.Cursor, re encoding.Resolution, shard, view uint64) *ro2.Bitmap {
	left := o.Left.Apply(cu, re, shard, view)
	return left.Union(o.Right.Apply(cu, re, shard, view))
}

type Yes struct {
	Field  models.Field
	Values []uint64
}

type No struct {
	Field  models.Field
	Values []uint64
}

// Apply searches for columns matching conditions in f for ra bitmap. ra must be
// mutex encoded.
func (f *Yes) Apply(cu *cursor.Cursor, re encoding.Resolution, shard, view uint64) *ro2.Bitmap {
	if len(f.Values) == 0 {
		return ro2.NewBitmap()
	}
	if !cu.ResetData(re, f.Field, view) {
		return ro2.NewBitmap()
	}
	all := make([]*ro2.Bitmap, 0, len(f.Values))
	for _, v := range f.Values {
		all = append(all, ro2.Row(cu, shard, v))
	}
	b := all[0]
	return b.Union(all[1:]...)
}

func (f *No) Apply(cu *cursor.Cursor, re encoding.Resolution, shard, view uint64) *ro2.Bitmap {
	if len(f.Values) == 0 {
		return ro2.NewBitmap()
	}
	if !cu.ResetExistence(re, f.Field, view) {
		return ro2.NewBitmap()
	}
	ex := ro2.Existence(cu, shard)
	if !cu.ResetData(re, f.Field, view) {
		return ro2.NewBitmap()
	}
	all := make([]*ro2.Bitmap, 0, len(f.Values))
	for _, v := range f.Values {
		all = append(all, ex.Difference(ro2.Row(cu, shard, v)))
	}
	b := all[0]
	return b.Union(all[1:]...)
}

func (ts *Timeseries) compile(fs wq.Filters) Filter {
	var a Filter
	for _, f := range fs {
		switch f.Key {
		case "city":
			switch f.Op {
			case "is", "is_not":
				code, _ := strconv.Atoi(f.Value[0])
				if code == 0 {
					return nil
				}
				value := []uint64{uint64(code)}
				if f.Op == "is" {
					a = And{
						Left: a,
						Right: &Yes{
							Field:  models.Field_city,
							Values: value,
						},
					}
				} else {
					a = And{
						Left: a,
						Right: &No{
							Field:  models.Field_city,
							Values: value,
						},
					}
				}
			}
		default:
			fd := models.Field(models.Field_value[f.Key])
			if fd == 0 {
				return nil
			}

			switch f.Op {
			case "is", "is_not":
				values := make([]uint64, len(f.Value))
				for i := range f.Value {
					values[i] = ts.Translate(fd, []byte(f.Value[i]))
				}
				if f.Op == "is" {
					a = And{
						Left: a,
						Right: &Yes{
							Field:  fd,
							Values: values,
						},
					}
				} else {
					a = And{
						Left: a,
						Right: &No{
							Field:  fd,
							Values: values,
						},
					}
				}
			case "matches", "does_not_match":
				var values []uint64
				for _, source := range f.Value {
					prefix, exact := searchPrefix([]byte(source))
					if exact {
						values = append(values, ts.Translate(fd, []byte(source)))
					} else {
						re, err := regexp.Compile(source)
						if err != nil {
							return nil
						}

						ts.Search(fd, prefix, func(key []byte, val uint64) {
							if re.Match(key) {
								values = append(values, val)
							}
						})
					}
				}
				if len(values) == 0 {
					return nil
				}

				if f.Op == "matches" {
					a = And{
						Left: a,
						Right: &Yes{
							Field:  fd,
							Values: values,
						},
					}
				} else {
					a = And{
						Left: a,
						Right: &No{
							Field:  fd,
							Values: values,
						},
					}
				}
			case "contains", "does_not_contain":
				var values []uint64
				re, err := regexp.Compile(strings.Join(f.Value, "|"))
				if err != nil {
					return nil
				}
				ts.Search(fd, []byte{}, func(b []byte, val uint64) {
					if re.Match(b) {
						values = append(values, val)
					}
				})

				if f.Op == "contains" {
					a = And{
						Left: a,
						Right: &Yes{
							Field:  fd,
							Values: values,
						},
					}
				} else {
					a = And{
						Left: a,
						Right: &No{
							Field:  fd,
							Values: values,
						},
					}
				}
			default:
				return nil
			}
		}
	}
	return a
}

func (ts *Timeseries) Search(field models.Field, prefix []byte, f func(key []byte, value uint64)) {
	sk := encoding.TranslateKey(field, prefix)
	data.Prefix(ts.db.Get(), sk, func(key, value []byte) error {
		f(key[3:], binary.BigEndian.Uint64(value))
		return nil
	})
}

func (ts *Timeseries) SearchKeys(field models.Field, prefix []byte, f func(key []byte) error) error {
	sk := encoding.TranslateKey(field, prefix)
	return data.PrefixKeys(ts.db.Get(), sk, func(key []byte) error {
		return f(key[3:])
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
