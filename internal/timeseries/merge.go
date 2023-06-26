package timeseries

import (
	"bytes"
	"context"
	"encoding/binary"
	"sync"

	"github.com/cespare/xxhash/v2"
	"github.com/dgraph-io/badger/v4"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/pkg/log"
)

type mergeFunction func(context.Context, uint64, *kvTs) error

func Merge(ctx context.Context) {
	mergeStats(ctx, joe)
}

func joe(ctx context.Context, ts uint64, kv *kvTs) error {
	return nil
}

func mergeStats(ctx context.Context, mergeFn mergeFunction) {
	ts := uint64(core.Now(ctx).UnixMilli())
	txn := GetMike(ctx).NewTransactionAt(ts, true)
	o := badger.DefaultIteratorOptions
	it := txn.NewIterator(o)
	ls := newTxnBufferList()
	defer ls.release()

	m := mergePool.Get().(*merge)
	defer m.release()

	for it.Rewind(); it.Valid(); it.Next() {
		x := it.Item()
		if x.IsDeletedOrExpired() {
			continue
		}
		key := x.Key()
		x.Value(func(val []byte) error {
			m.add(key, val)
			return nil
		})
		k := ls.Get()
		k.Write(key)
		txn.Delete(k.Bytes())
	}
	err := txn.Commit()
	if err != nil {
		log.Get().Err(err).Msg("failed to commit merge transaction")
		return
	}
	err = m.do(ctx, mergeFn)
	if err != nil {
		log.Get().Err(err).Msg("failed merge operation")
	}
}

type merge struct {
	ts map[uint64]*kvTs
	m  map[uint64]*kvBuf
	h  *xxhash.Digest
}

func (m *merge) release() {
	for k, v := range m.m {
		v.reset()
		delete(m.m, k)
	}
	for k, v := range m.ts {
		v.reset()
		delete(m.ts, k)
	}
	m.h.Reset()
	mergePool.Put(m)
}

var mergePool = &sync.Pool{
	New: func() any {
		return &merge{
			ts: make(map[uint64]*kvTs),
			m:  make(map[uint64]*kvBuf),
			h:  xxhash.New(),
		}
	},
}

func (m *merge) hash(b []byte) uint64 {
	m.h.Reset()
	m.h.Write(b)
	return m.h.Sum64()
}

func (m *merge) add(key, value []byte) {
	m.h.Reset()
	baseKey := key[:len(key)-8]
	baseTs := key[len(key)-8:]
	keyHash := m.hash(baseKey)
	ts := binary.BigEndian.Uint64(baseTs)
	m.h.Reset()
	v := binary.BigEndian.Uint32(value)
	b, ok := m.m[keyHash]
	if ok {
		// existing key
		b.b = append(b.b, v)
	} else {
		b = kvBufPool.Get().(*kvBuf)
		b.k = get()
		b.k.Write(baseKey)
		m.m[keyHash] = b
		t, ok := m.ts[ts]
		if !ok {
			t = kvTsPool.Get().(*kvTs)
			m.ts[ts] = t
		}
		t.b = append(t.b, b)
	}
}

func (m *merge) do(ctx context.Context, f mergeFunction) error {
	for k, kt := range m.ts {
		err := f(ctx, k, kt)
		if err != nil {
			return err
		}
	}
	return nil
}

type kvTs struct {
	b []*kvBuf
}

func (k *kvTs) reset() {
	k.b = k.b[:0]
	kvTsPool.Put(k)
}

var kvTsPool = &sync.Pool{
	New: func() any {
		return &kvTs{
			b: make([]*kvBuf, 0, 1<<10),
		}
	},
}

type kvBuf struct {
	k *bytes.Buffer
	b []uint32
}

func (k *kvBuf) reset() {
	put(k.k)
	k.b = k.b[:0]
	kvBufPool.Put(k)
}

var kvBufPool = &sync.Pool{
	New: func() any {
		return &kvBuf{
			b: make([]uint32, 0, 1<<10),
		}
	},
}

type kv struct {
	k *bytes.Buffer
	v uint32
}
