package ro2

import (
	"hash/crc32"
	"slices"
	"time"

	"github.com/vinceanalytics/vince/internal/ro"
	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
)

// all fields
const (
	domainField    uint64 = 13
	timestampField        = 1
	idField               = 2
	bounceField           = 3
	sessionField          = 4
	viewField             = 5
	durationField         = 7
	cityField             = 8
	eventsField           = 31
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
			ts := tx.Cmp(timestampField, shard, 64, roaring64.RANGE, start, end)
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

type Eq struct {
	field uint64
	value uint64
}

func newEq(field uint64, value string) *Eq {
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
