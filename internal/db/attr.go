package db

import (
	"encoding/binary"
	"errors"
	"hash/maphash"
	"os"
	"path/filepath"
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/gernest/rbf"
	"github.com/gernest/rbf/dsl/boolean"
	"github.com/gernest/rbf/dsl/bsi"
	"github.com/gernest/rbf/dsl/cursor"
	"github.com/gernest/rbf/dsl/mutex"
	"github.com/gernest/roaring"
	"github.com/gernest/roaring/shardwidth"
	"github.com/gernest/rows"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"go.etcd.io/bbolt"
)

var (
	idsBucket = []byte("\x00i")
	trKey     = []byte("\x00tk")
	trID      = []byte("\x00ti")
)

func (db *DB) view(f func(tx *view) error) error {
	tx, err := db.idx.Begin(false)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txn, err := db.db.Begin(false)
	if err != nil {
		return err
	}
	defer txn.Rollback()
	vx := &view{
		tx: tx, txn: txn, m: make(map[string]*rbf.Cursor),
	}
	defer vx.Release()
	return f(vx)
}

func (db *DB) Save() error {
	b := db.batch
	defer b.Reset()
	return db.db.Update(func(tx *bbolt.Tx) error {
		idx, err := bucket(tx, idsBucket)
		if err != nil {
			return err
		}

		m := map[string]*roaring.Bitmap{}

		tr := newTr(tx)
		var id uint64
		for i := range b.ts {
			id, err = idx.NextSequence()
			if err != nil {
				return err
			}
			mutex.Add(get(m, "_id"), id, id)
			bsi.Add(get(m, "timestamp"), id, b.ts[i])
			bsi.Add(get(m, "date"), id, date(b.ts[i]))
			bsi.Add(get(m, "uid"), id, b.uid[i])
			boolean.Add(get(m, "bounce"), id, b.bounce[i])
			boolean.Add(get(m, "session"), id, b.session[i])
			boolean.Add(get(m, "view"), id, b.view[i])
			bsi.Add(get(m, "duration"), id, b.duration[i])
			for k, v := range b.attr[i] {
				x, err := tr.Tr(k, v)
				if err != nil {
					return err
				}
				mutex.Add(get(m, k), id, x)
			}
		}
		shard := id / shardwidth.ShardWidth

		txn, err := db.idx.Begin(true)
		if err != nil {
			return err
		}
		for k, v := range m {
			_, err := txn.AddRoaring(k, v)
			if err != nil {
				txn.Rollback()
				return err
			}
		}
		err = txn.Commit()
		if err != nil {
			return err
		}
		if db.shards.Contains(shard) {
			return nil
		}

		// update SHARDS field
		db.shards.Add(shard)
		db.shards.RunOptimize()
		data, _ := db.shards.MarshalBinary()
		return os.WriteFile(filepath.Join(db.idx.Path, "SHARDS"), data, 0600)
	})
}

func date(ts int64) int64 {
	yy, mm, dd := time.UnixMilli(ts).Date()
	return time.Date(yy, mm, dd, 0, 0, 0, 0, time.UTC).UnixMilli()
}

func get(m map[string]*roaring.Bitmap, key string) *roaring.Bitmap {
	if b, ok := m[key]; ok {
		return b
	}
	b := roaring.NewBitmap()
	m[key] = b
	return b
}

func bucket(tx *bbolt.Tx, key []byte) (*bbolt.Bucket, error) {
	if b := tx.Bucket(key); b != nil {
		return b, nil
	}
	return tx.CreateBucket(key)
}

func bbucket(tx *bbolt.Bucket, key []byte) (*bbolt.Bucket, error) {
	if b := tx.Bucket(key); b != nil {
		return b, nil
	}
	return tx.CreateBucket(key)
}

type translate struct {
	tx      *bbolt.Tx
	buckets map[string]*bbolt.Bucket
	seen    map[uint64]uint64
	h       maphash.Hash
}

func newTr(tx *bbolt.Tx) *translate {
	return &translate{
		tx:      tx,
		buckets: make(map[string]*bbolt.Bucket),
		seen:    make(map[uint64]uint64),
	}
}

var sep = []byte("=")

func (tr *translate) Tr(key, value string) (v uint64, err error) {
	tr.h.Reset()
	tr.h.WriteString(key)
	tr.h.Write(sep)
	tr.h.WriteString(value)
	hash := tr.h.Sum64()
	if v, ok := tr.seen[hash]; ok {
		return v, nil
	}
	b, ok := tr.buckets[key]
	if !ok {
		b, err = bucket(tr.tx, []byte(key))
		if err != nil {
			return 0, err
		}
		tr.buckets[key] = b
	}
	rk := append(trKey, []byte(value)...)
	if x := b.Get(rk); x != nil {
		v = binary.BigEndian.Uint64(x)
		tr.seen[hash] = v
		return v, nil
	}
	nxt, err := b.NextSequence()
	if err != nil {
		return 0, err
	}
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], nxt)
	ri := append(trID, buf[:]...)
	err = b.Put(ri, []byte(value))
	if err != nil {
		return 0, err
	}
	err = b.Put(rk, buf[:])
	if err != nil {
		return 0, err
	}
	tr.seen[hash] = nxt
	return nxt, nil
}

func eq(tx *view, shard uint64, key, value string) (*rows.Row, error) {
	c, err := tx.get(key)
	if err != nil {
		if errors.Is(err, rbf.ErrBitmapNotFound) {
			return rows.NewRow(), nil
		}
		return nil, err
	}
	defer c.Close()
	row, ok := tx.find(key, value)
	if !ok {
		return rows.NewRow(), nil
	}
	return cursor.Row(c, shard, row)
}

func find(tx *bbolt.Tx, key, value string) (uint64, bool) {
	b := tx.Bucket([]byte(key))
	if b == nil {
		return 0, false
	}
	id := b.Get(append(trKey, []byte(value)...))
	data := b.Get(id)
	if len(data) != 0 {
		return binary.BigEndian.Uint64(data), true
	}
	return 0, false
}

func (tx *view) time(shard uint64, start, end time.Time, r *rows.Row) (*rows.Row, error) {
	ts, err := tx.get("timestamp")
	if err != nil {
		return nil, err
	}
	return bsi.Compare(ts, shard, bsi.RANGE, start.UnixMilli(), end.UnixMilli(), r)
}

func (tx *view) domain(shard uint64, name string) (*rows.Row, error) {
	return eq(tx, shard, v1.Property_domain.String(), name)
}

func (tx *view) boolCount(field string, shard uint64, isTrue bool, columns *rows.Row) (count int64, err error) {
	c, err := tx.get(field)
	if err != nil {
		return 0, err
	}
	var r *rows.Row
	if isTrue {
		r, err = cursor.Row(c, shard, 1)
	} else {
		r, err = cursor.Row(c, shard, 0)
	}
	if err != nil {
		return 0, err
	}
	if columns != nil {
		r = r.Intersect(columns)
	}
	count = int64(r.Count())
	return
}

const (
	shardVsContainerExponent = shardwidth.Exponent - 16
)

func (tx *view) distinct(field string, o *roaring64.Bitmap, filter []*roaring.Container, isFilter bool) error {
	c, err := tx.get(field)
	if err != nil {
		return err
	}
	fragData := c.Iterator()

	prevRow := ^uint64(0)
	seenThisRow := false
	for fragData.Next() {
		k, c := fragData.Value()
		row := k >> shardVsContainerExponent
		if row == prevRow {
			if seenThisRow {
				continue
			}
		} else {
			seenThisRow = false
			prevRow = row
		}
		if isFilter {
			if roaring.IntersectionAny(c, filter[k%(1<<shardVsContainerExponent)]) {
				o.Add(row)
				seenThisRow = true
			}
		} else if c.N() != 0 {
			o.Add(row)
			seenThisRow = true
		}
	}
	return nil
}

func (tx *view) row(field string, shard uint64, id uint64, f *rows.Row) (*rows.Row, error) {
	c, err := tx.get(field)
	if err != nil {
		return nil, err
	}
	r, err := cursor.Row(c, shard, id)
	if err != nil {
		return nil, err
	}
	return r.Intersect(f), nil

}

func (tx *view) key(field string, id uint64) []byte {
	b := tx.txn.Bucket([]byte(field))
	if b == nil {
		return []byte{}
	}
	var a [8]byte
	binary.BigEndian.PutUint64(a[:], id)
	return b.Get(append(trID, a[:]...))
}

func (tx *view) unique(field string, shard uint64,
	columns *rows.Row) (o map[uint64][]uint64, err error) {
	c, err := tx.get(field)
	if err != nil {
		return nil, err
	}
	exists, err := cursor.Row(c, shard, 0)
	if err != nil {
		return nil, err
	}
	exists = exists.Intersect(columns)
	if exists.IsEmpty() {
		return nil, nil
	}
	m := make(map[uint64]uint64)
	mergeBits(m, exists.Columns(), 0)

	for i := uint64(0); i < 64; i++ {
		bits, err := cursor.Row(c, shard, 2+uint64(i))
		if err != nil {
			return nil, err
		}
		if bits.IsEmpty() {
			continue
		}
		bits = bits.Intersect(exists)
		if bits.IsEmpty() {
			continue
		}
		mergeBits(m, bits.Columns(), 1<<i)
	}
	o = make(map[uint64][]uint64, len(m))
	for column, val := range m {
		// Convert to two's complement and add base back to value.
		val = uint64((2*(int64(val)>>63) + 1) * int64(val&^(1<<63)))
		o[val] = append(o[val], column)
	}
	return
}

func mergeBits(m map[uint64]uint64, bits []uint64, mask uint64) {
	for _, v := range bits {
		m[v] |= mask
	}
}
