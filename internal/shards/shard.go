package shards

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/ro2"
	"github.com/vinceanalytics/vince/internal/util/assert"
	"github.com/vinceanalytics/vince/internal/util/data"
)

func Format(shard uint64) string {
	return fmt.Sprintf("%006d", shard)
}

func Parse(str string) uint64 {
	a, err := strconv.ParseUint(str, 10, 64)
	assert.Nil(err)
	return a
}

type Views [encoding.Month + 1]*ro2.Bitmap

func (v *Views) Init() {
	for i := range v {
		v[i] = ro2.NewBitmap()
	}
}

func (v *Views) Reset() {
	for i := range v {
		v[i].Containers.Reset()
	}
}

type DB struct {
	ops    *pebble.DB
	base   string
	shards struct {
		sync.RWMutex
		data [1 << 10]*Shard
		max  uint64
	}
}

func New(base string) (*DB, error) {
	ops, err := data.Open(filepath.Join(base, "ops"), nil)
	if err != nil {
		return nil, err
	}
	db := &DB{ops: ops, base: base}
	shardsBase := filepath.Join(base, "shards")
	entries, _ := os.ReadDir(shardsBase)
	shards := make([]uint64, 0, len(entries))
	for _, e := range entries {
		shards = append(shards, Parse(e.Name()))
	}
	slices.Sort(shards)
	for _, sh := range shards {
		db.shards.data[sh] = NewShard(shardsBase, sh)
		db.shards.max = max(db.shards.max, sh)
	}
	return db, nil
}

func (db *DB) Get() *pebble.DB {
	return db.ops
}

func (db *DB) Iter(re encoding.Resolution, start, end time.Time, f func(db *pebble.DB, shard uint64, views *ro2.Bitmap) error) error {
	db.shards.RLock()
	defer db.shards.RUnlock()
	a := uint64(start.UnixMilli())
	b := uint64(end.UnixMilli())
	for i := uint64(0); i <= db.shards.max; i++ {
		sh := db.shards.data[i]
		if sh == nil {
			continue
		}

		o := sh.ComputeViews(re, a, b)

		if o.Any() {
			err := f(sh.DB, sh.ID, o)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *DB) Shard(shard uint64) *Shard {
	db.shards.RLock()
	sh := db.shards.data[shard]
	db.shards.RUnlock()
	if sh != nil {
		return sh
	}
	db.shards.Lock()
	defer db.shards.Unlock()
	sh = NewShard(filepath.Join(db.base, "shards"), shard)
	db.shards.data[shard] = sh
	db.shards.max = max(db.shards.max, shard)
	return sh
}

func (db *DB) Close() error {
	var errs []error
	errs = append(errs, db.ops.Close())
	db.shards.Lock()
	defer db.shards.Unlock()

	for _, sh := range db.shards.data {
		if sh != nil {
			errs = append(errs, sh.DB.Close())
		}
	}
	return errors.Join(errs...)
}

type Shard struct {
	ID    uint64
	DB    *pebble.DB
	views struct {
		sync.RWMutex
		data Views
	}
}

func NewShard(base string, shard uint64) *Shard {
	path := filepath.Join(base, Format(shard))
	db, err := data.Open(path, &pebble.Options{
		Merger: ro2.Merge,
	})
	assert.Nil(err, fmt.Sprintf("opening database shard path=%q", path))
	s := &Shard{ID: shard, DB: db}
	s.views.data.Init()
	err = data.Prefix(db, keys.ViewsPrefix, func(key, value []byte) error {
		r := key[len(key)-1]
		return s.views.data[r].UnmarshalBinary(value)
	})
	assert.Nil(err, "loading views")
	return s
}

func (sh *Shard) UpdateViews(vs *Views, ba *pebble.Batch) error {
	sh.views.Lock()
	for i := range vs {
		sh.views.data[i].UnionInPlace(vs[i])
	}
	sh.views.Unlock()

	var b bytes.Buffer
	sh.views.RLock()
	defer sh.views.RUnlock()
	key := make([]byte, 2)
	copy(key, keys.ViewsPrefix)
	for i := range sh.views.data {
		b.Reset()
		_, err := sh.views.data[i].WriteTo(&b)
		if err != nil {
			return fmt.Errorf("marshal views bitmap %w", err)
		}
		key[len(key)-1] = byte(i)
		err = ba.Set(key, b.Bytes(), nil)
		if err != nil {
			return fmt.Errorf("saving view bitmap %w", err)
		}
	}
	return nil
}

func (sh *Shard) ComputeViews(re encoding.Resolution, start, end uint64) *ro2.Bitmap {
	if re == encoding.Global {
		o := ro2.NewBitmap()
		o.Add(0)
		return o
	}
	sh.views.RLock()
	defer sh.views.RUnlock()
	return rangeBits(sh.views.data[re], start, end)
}

func rangeBits(b *ro2.Bitmap, start, end uint64) *ro2.Bitmap {
	itr := b.Iterator()
	itr.Seek(start)
	o := ro2.NewBitmap()
	for v, eof := itr.Next(); !eof && v <= end; v, eof = itr.Next() {
		o.Add(v)
	}
	return o
}
