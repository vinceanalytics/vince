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

// Cond defines exact matches rows and non exact match rows. Applies a union of
// all columns satsfying the conditions.
//
// This assumes Cond is for a mutex field. vince only support filter on mutex
// fields.
type Cond struct {
	Yes []uint64
	No  []uint64
}

// IsEmpty return true if there is no row in yes or no conditions.
func (f *Cond) IsEmpty() bool {
	return len(f.Yes) == 0 && len(f.No) == 0
}

// Apply searches for columns matching conditions in f for ra bitmap. ra must be
// mutex encoded.
func (f *Cond) Apply(shard uint64, cu *cursor.Cursor, exists *ro2.Bitmap) *ro2.Bitmap {
	if f.IsEmpty() {
		return ro2.NewBitmap()
	}
	all := make([]*ro2.Bitmap, 0, len(f.Yes)+len(f.No))

	for _, v := range f.Yes {
		all = append(all, ro2.Row(cu, shard, v))
	}
	for _, v := range f.No {
		all = append(all, exists.Difference(ro2.Row(cu, shard, v)))
	}
	b := all[0]
	return b.Union(all[1:]...)
}

type FilterSet [models.SearchFieldSize]Cond

// ScanFields returns a set of all fields to scan for this filter.
func (fs *FilterSet) ScanFields() (set models.BitSet) {
	fs.idx(func(i int) {
		set.Set(models.Field(i))
	})
	return set
}

func (fs *FilterSet) idx(f func(int)) {
	for i := range fs {
		if fs[i].IsEmpty() {
			continue
		}
		f(i)
	}
}

func (fs *FilterSet) Set(yes bool, f models.Field, values ...uint64) {
	co := &fs[f]
	if yes {
		co.Yes = append(co.Yes, values...)
		return
	}
	co.No = append(co.No, values...)
}

func (ts *Timeseries) compile(fs wq.Filters) FilterSet {
	var a FilterSet
	for _, f := range fs {
		switch f.Key {
		case "city":
			switch f.Op {
			case "is", "is_not":
				code, _ := strconv.Atoi(f.Value[0])
				if code == 0 {
					return FilterSet{}
				}
				a.Set(f.Op == "is", models.Field_city, uint64(code))
			}
		default:
			fd := models.Field(models.Field_value[f.Key])
			if fd == 0 {
				return FilterSet{}
			}

			switch f.Op {
			case "is", "is_not":
				values := make([]uint64, len(f.Value))
				for i := range f.Value {
					values[i] = ts.Translate(fd, []byte(f.Value[i]))
				}
				a.Set(f.Op == "is", fd, values...)
			case "matches", "does_not_match":
				var values []uint64
				for _, source := range f.Value {
					prefix, exact := searchPrefix([]byte(source))
					if exact {
						values = append(values, ts.Translate(fd, []byte(source)))
					} else {
						re, err := regexp.Compile(source)
						if err != nil {
							return FilterSet{}
						}

						ts.Search(fd, prefix, func(key []byte, val uint64) {
							if re.Match(key) {
								values = append(values, val)
							}
						})
					}
				}
				if len(values) == 0 {
					return FilterSet{}
				}

				a.Set(f.Op == "matches", fd, values...)

			case "contains", "does_not_contain":
				var values []uint64
				re, err := regexp.Compile(strings.Join(f.Value, "|"))
				if err != nil {
					return FilterSet{}
				}
				ts.Search(fd, []byte{}, func(b []byte, val uint64) {
					if re.Match(b) {
						values = append(values, val)
					}
				})

				a.Set(f.Op == "contains", fd, values...)

			default:
				return FilterSet{}
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
