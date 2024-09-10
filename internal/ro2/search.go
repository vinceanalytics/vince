package ro2

import (
	"encoding/binary"
	"log/slog"
	"regexp"
	"slices"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/vinceanalytics/vince/internal/alicia"
	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
)

// We know fields before hand
type Data [alicia.SUB2_CODE]*roaring64.BSI

var dataPool = &sync.Pool{
	New: func() any {
		var d Data
		return &d
	},
}

func NewData() *Data {
	return dataPool.Get().(*Data)
}

func (d *Data) Release() {
	clear(d[:])
	dataPool.Put(d)
}

func (d *Data) get(i alicia.Field) *roaring64.BSI {
	i--
	if d[i] == nil {
		d[i] = roaring64.NewDefaultBSI()
	}
	return d[i]
}

func (o *Store) Select(
	start, end int64,
	domain string,
	filter Filter,
	f func(tx *Tx, shard uint64, match *roaring64.Bitmap) error) error {

	if filter == nil {
		filter = noop{}
	}

	shards := o.shards()
	if len(shards) == 0 {
		return nil
	}

	dom := NewEq(uint64(alicia.DOMAIN), domain)
	return o.View(func(tx *Tx) error {

		// We iterate on shards in reverse. We are always interested in latest data.
		// This way we can have early exit when the caller is done but we have more
		// shards left.
		slices.Reverse(shards)

		for i := range shards {
			shard := shards[i]

			b := dom.match(tx, shard)
			if b.IsEmpty() {
				continue
			}

			filter.apply(tx, shard, b)
			if b.IsEmpty() {
				continue
			}

			// select timestamp time range
			ts := tx.Cmp(uint64(alicia.TIMESTAMP), shard, roaring64.RANGE, start, end)
			b.And(ts)
			if b.IsEmpty() {
				continue
			}

			err := f(tx, shard, b)
			if err != nil {
				return err
			}
		}
		return nil
	})

}

type Filter interface {
	apply(tx *Tx, shard uint64, match *roaring64.Bitmap)
}

type List []Filter

func (ls List) apply(tx *Tx, shard uint64, match *roaring64.Bitmap) {
	for i := range ls {
		if match.IsEmpty() {
			return
		}
		ls[i].apply(tx, shard, match)
	}
}

type noop struct{}

func (noop) apply(tx *Tx, shard uint64, match *roaring64.Bitmap) {}

type Reject struct{}

func (Reject) apply(tx *Tx, shard uint64, match *roaring64.Bitmap) { match.Clear() }

type Regex struct {
	field uint64
	value *regexp.Regexp
	valid roaring64.Bitmap
	once  sync.Once
}

func NewRe(field uint64, value string) Filter {
	r, err := regexp.Compile(cleanRe(value))
	if err != nil {
		slog.Error("invalid regex filter", "field", field, "value", value)
		return Reject{}
	}
	return &Regex{
		field: field,
		value: r,
	}
}

func (e *Regex) apply(tx *Tx, shard uint64, match *roaring64.Bitmap) {
	match.And(e.match(tx, shard))
}

func (e *Regex) match(tx *Tx, shard uint64) *roaring64.Bitmap {
	e.once.Do(func() {
		source := e.value.String()
		prefix, exact := searchPrefix([]byte(source))
		if exact {
			// fast path. We only perform a single Get call
			key := tx.get().TranslateKey(e.field, prefix)
			it, err := tx.tx.Get(key)
			if err == nil {
				it.Value(func(val []byte) error {
					e.valid.Add(binary.BigEndian.Uint64(val))
					return nil
				})
			}
		} else {
			tx.Search(e.field, prefix, func(b []byte, u uint64) {
				e.valid.Add(u)
			})
		}
	})
	if e.valid.IsEmpty() {
		return &e.valid
	}
	union := roaring64.New()

	it := e.valid.Iterator()
	for it.HasNext() {
		union.Or(
			tx.Cmp(e.field, shard, roaring64.EQ, int64(it.Next()), 0),
		)
	}
	return union
}

type Nre struct {
	field uint64
	value *regexp.Regexp
	valid roaring64.Bitmap
	once  sync.Once
}

func NewNre(field uint64, value string) Filter {
	r, err := regexp.Compile(cleanRe(value))
	if err != nil {
		slog.Error("invalid regex filter", "field", field, "value", value)
		return Reject{}
	}
	return &Nre{
		field: field,
		value: r,
	}
}

func (e *Nre) apply(tx *Tx, shard uint64, match *roaring64.Bitmap) {
	match.And(e.match(tx, shard))
}

func (e *Nre) match(tx *Tx, shard uint64) *roaring64.Bitmap {
	e.once.Do(func() {
		source := e.value.String()
		prefix, exact := searchPrefix([]byte(source))
		if exact {
			// fast path. We only perform a single Get call
			key := tx.get().TranslateKey(e.field, prefix)
			it, err := tx.tx.Get(key)
			if err == nil {
				it.Value(func(val []byte) error {
					e.valid.Add(binary.BigEndian.Uint64(val))
					return nil
				})
			}
		} else {
			tx.Search(e.field, prefix, func(b []byte, u uint64) {
				e.valid.Add(u)
			})
		}
	})
	if e.valid.IsEmpty() {
		return &e.valid
	}
	union := roaring64.New()

	it := e.valid.Iterator()
	for it.HasNext() {
		union.Or(
			tx.Cmp(e.field, shard, roaring64.NEQ, int64(it.Next()), 0),
		)
	}
	return union
}

type Eq struct {
	field uint64
	value string
	id    uint64
	once  sync.Once
}

func NewEq(field uint64, value string) *Eq {
	return &Eq{
		field: field,
		value: value,
	}
}

func (e *Eq) apply(tx *Tx, shard uint64, match *roaring64.Bitmap) {
	match.And(e.match(tx, shard))
}

func (e *Eq) match(tx *Tx, shard uint64) *roaring64.Bitmap {
	e.once.Do(func() {
		key := tx.get().TranslateKey(e.field, []byte(e.value))
		it, err := tx.tx.Get(key)
		if err == nil {
			it.Value(func(val []byte) error {
				e.id = binary.BigEndian.Uint64(val)
				return nil
			})
		}
	})
	if e.id == 0 {
		return roaring64.New()
	}
	return tx.Cmp(e.field, shard, roaring64.EQ, int64(e.id), 0)
}

type Neq struct {
	field uint64
	value string
	id    uint64
	once  sync.Once
}

func NewNeq(field uint64, value string) *Neq {
	return &Neq{
		field: field,
		value: value,
	}
}

func (e *Neq) apply(tx *Tx, shard uint64, match *roaring64.Bitmap) {
	match.And(e.match(tx, shard))
}

func (e *Neq) match(tx *Tx, shard uint64) *roaring64.Bitmap {
	e.once.Do(func() {
		key := tx.get().TranslateKey(e.field, []byte(e.value))
		it, err := tx.tx.Get(key)
		if err == nil {
			it.Value(func(val []byte) error {
				e.id = binary.BigEndian.Uint64(val)
				return nil
			})
		}
	})
	if e.id == 0 {
		return roaring64.New()
	}
	return tx.Cmp(e.field, shard, roaring64.NEQ, int64(e.id), 0)
}

type EqInt struct {
	Field uint64
	Value int64
}

func (e *EqInt) apply(tx *Tx, shard uint64, match *roaring64.Bitmap) {
	match.And(
		tx.Cmp(e.Field, shard, roaring64.EQ, e.Value, 0),
	)
}

func cleanRe(re string) string {
	re = strings.TrimPrefix(re, "~")
	re = strings.TrimSuffix(re, "$")
	return re
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
