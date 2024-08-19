package ro2

import (
	"hash/crc32"
	"log/slog"
	"regexp"
	"slices"
	"time"

	"github.com/vinceanalytics/vince/internal/ro"
	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
)

// all fields
const (
	timestampField         = 1
	idField                = 2
	bounceField            = 3
	sessionField           = 4
	viewField              = 5
	durationField          = 7
	cityField              = 8
	BrowserField           = 9
	Browser_versionField   = 10
	CountryField           = 11
	DeviceField            = 12
	domainField            = 13
	Entry_pageField        = 14
	eventField             = 15
	exit_pageField         = 16
	hostField              = 17
	OsField                = 18
	Os_versionField        = 19
	PageField              = 20
	ReferrerField          = 21
	SourceField            = 22
	Utm_campaignField      = 23
	Utm_contentField       = 24
	Utm_mediumField        = 25
	Utm_sourceField        = 26
	Utm_termField          = 27
	Subdivision1_codeField = 28
	subdivision2_codeField = 29
	eventsField            = 31
)

// We know fields before hand
type Data [32]*roaring64.BSI

func (d *Data) get(i uint32) *roaring64.BSI {
	if d[i] == nil {
		d[i] = roaring64.NewDefaultBSI()
	}
	return d[i]
}

type Match struct {
	Dates []uint64
	Data  []Data
}

func (m *Match) Reset(dates []uint64) {
	m.Dates = slices.Grow(m.Dates, len(dates))[:len(dates)]
	copy(m.Dates, dates)
	m.Data = slices.Grow(m.Data, len(dates))[:len(dates)]
	for i := range m.Data {
		clear(m.Data[i][:])
	}
}

func (o *Proto[T]) Select(
	start, end int64,
	domain string,
	filter Filter,
	f func(tx *Tx, shard uint64, match *roaring64.Bitmap) error) error {

	dates := ro.DateRange(
		time.UnixMilli(start).UTC(),
		time.UnixMilli(end).UTC(),
	)
	if len(dates) == 0 {
		return nil
	}

	if filter == nil {
		filter = noop{}
	}

	shards := o.shards()

	hash := crc32.NewIEEE()
	hash.Write([]byte(domain))
	domainID := uint64(hash.Sum32())

	return o.View(func(tx *Tx) error {

		// We iterate on shards in reverse. We are always interested in latest data.
		// This way we can have early exit when the caller is done but we have more
		// shards left.
		slices.Reverse(shards)

		for i := range shards {
			shard := shards[i]

			b := tx.Row(shard, domainField, domainID)
			if b.IsEmpty() {
				continue
			}

			filter.apply(tx, shard, b)
			if b.IsEmpty() {
				continue
			}

			// select timestamp
			ts := tx.Cmp(timestampField, shard, roaring64.RANGE, start, end)
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

// compute possible shards based on current id
func (o *Proto[T]) shards() []uint64 {
	shard := o.seq.Load() / ro.ShardWidth
	n := make([]uint64, 0, shard+1)
	for i := uint64(0); i <= shard; i++ {
		n = append(n, i)
	}
	return n
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
}

func NewRe(field uint64, value string) Filter {
	r, err := regexp.Compile(value)
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
	hash := crc32.NewIEEE()
	union := roaring64.New()
	tx.searchTranslation(shard, e.field, func(val []byte) {
		if e.value.Match(val) {
			hash.Reset()
			hash.Write(val)
			union.Or(
				tx.Row(shard, e.field, uint64(hash.Sum32())),
			)
		}
	})
	match.And(union)
}

type Nre struct {
	field uint64
	value *regexp.Regexp
}

func NewNre(field uint64, value string) Filter {
	r, err := regexp.Compile(value)
	if err != nil {
		slog.Error("invalid regex filter", "field", field, "value", value)
		return Reject{}
	}
	return &Regex{
		field: field,
		value: r,
	}
}

func (e *Nre) apply(tx *Tx, shard uint64, match *roaring64.Bitmap) {
	hash := crc32.NewIEEE()
	union := roaring64.New()
	tx.searchTranslation(shard, e.field, func(val []byte) {
		if e.value.Match(val) {
			hash.Reset()
			hash.Write(val)
			union.Or(
				tx.Row(shard, e.field, uint64(hash.Sum32())),
			)
		}
	})
	exists := tx.Row(shard, timestampField, 0)
	match.And(roaring64.AndNot(exists, union))
}

type Eq struct {
	field uint64
	value uint64
}

func NewEq(field uint64, value string) *Eq {
	hash := crc32.NewIEEE()
	hash.Write([]byte(value))
	return &Eq{
		field: field,
		value: uint64(hash.Sum32()),
	}
}

func (e *Eq) apply(tx *Tx, shard uint64, match *roaring64.Bitmap) {
	match.And(
		tx.Row(shard, e.field, e.value),
	)
}

type Neq struct {
	field uint64
	value uint64
}

func NewNeq(field uint64, value string) *Eq {
	hash := crc32.NewIEEE()
	hash.Write([]byte(value))
	return &Eq{
		field: field,
		value: uint64(hash.Sum32()),
	}
}

func (e *Neq) apply(tx *Tx, shard uint64, match *roaring64.Bitmap) {
	// we know timestamp is always set and the it is bsi. In the current shard it
	// is the best fiend to look for existence bitmap.
	exists := tx.Row(shard, timestampField, 0)
	match.And(
		roaring64.AndNot(exists, tx.Row(shard, e.field, e.value)),
	)
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
