package timeseries

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/gernest/vince/pkg/log"
	"github.com/gernest/vince/system"
)

type Sum struct {
	Visitors      uint32
	Views         uint32
	Events        uint32
	Visits        uint32
	BounceRate    uint32
	VisitDuration uint32
	ViewsPerVisit uint32
}

func DropSite(ctx context.Context, uid, sid uint64) {
	start := time.Now()
	defer system.DropSiteDuration.Observe(time.Since(start).Seconds())

	db := GetMike(ctx)
	id := newMetaKey()
	defer id.Release()
	id.uid(uid, sid)
	// remove all keys under /user_id/site_id/ prefix.
	err := db.DropPrefix(id[:metricOffset])
	if err != nil {
		log.Get().Err(err).
			Uint64("uid", uid).
			Uint64("sid", sid).
			Msg("failed to delete site from stats storage")
	}

}

func Save(ctx context.Context, b *Buffer) {
	start := time.Now()

	db := GetMike(ctx)
	meta := newMetaKey()
	ls := newTxnBufferList()

	svc := saveContext{}

	defer func() {
		log.Get().Debug().Int(
			"__size__", len(b.segments.Timestamp),
		).Int("__keys__", svc.keys).
			Msg("saved stats")
		b.Release()
		meta.Release()
		ls.Release()
		system.SaveDuration.Observe(time.Since(start).Seconds())
	}()
	svc.ls = ls

	// Buffer.id has the same encoding as the first 16 bytes of meta. We just copy that
	// there is no need to re encode user id and site id.
	copy(meta[:], b.id[:])
	var tsBytes [6]byte
	setTs(tsBytes[:], uint64(start.UnixMilli()))

	err := db.Update(func(txn *badger.Txn) error {
		svc.txn = txn
		return b.Build(ctx, func(p Property, key string, sum *Sum) error {
			return saveProperty(ctx, &svc, tsBytes[:], meta.prop(p), key, sum)
		})
	})
	if err != nil {
		log.Get().Err(err).Msg("failed to save ts buffer")
	}
}

func saveProperty(ctx context.Context,
	svc *saveContext,
	ts []byte,
	m *Key, text string, a *Sum) error {
	return errors.Join(
		savePropertyKey(ctx, svc, m.metric(Visitors).key(ts, text, svc.ls), a.Visitors),
		savePropertyKey(ctx, svc, m.metric(Views).key(ts, text, svc.ls), a.Views),
		savePropertyKey(ctx, svc, m.metric(Events).key(ts, text, svc.ls), a.Events),
		savePropertyKey(ctx, svc, m.metric(Visits).key(ts, text, svc.ls), a.Visits),
		savePropertyKey(ctx, svc, m.metric(BounceRates).key(ts, text, svc.ls), a.BounceRate),
		savePropertyKey(ctx, svc, m.metric(VisitDurations).key(ts, text, svc.ls), a.VisitDuration),
		savePropertyKey(ctx, svc, m.metric(ViewsPerVisits).key(ts, text, svc.ls), a.ViewsPerVisit),
	)
}

type saveContext struct {
	txn  *badger.Txn
	ls   *txnBufferList
	keys int
}

// Transaction keep reference to keys. We need this to make sure we reuse ID and properly
// release them back to the pool when done.
type txnBufferList struct {
	ls []*bytes.Buffer
}

func newTxnBufferList() *txnBufferList {
	return txnBufferListPool.New().(*txnBufferList)
}

func (ls *txnBufferList) Get() *bytes.Buffer {
	b := smallBufferpool.Get().(*bytes.Buffer)
	ls.ls = append(ls.ls, b)
	return b
}

func (ls *txnBufferList) Reset() {
	for _, v := range ls.ls {
		v.Reset()
		smallBufferpool.Put(v)
	}
	ls.ls = ls.ls[:0]
}

func (ls *txnBufferList) Release() {
	ls.Reset()
	txnBufferListPool.Put(ls)
}

var txnBufferListPool = &sync.Pool{
	New: func() any {
		return &txnBufferList{
			ls: make([]*bytes.Buffer, 0, 1<<10),
		}
	},
}

var smallBufferpool = &sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

// Use this to grow bytes.Buffer so we can use it in transaction value. This allows
// us to ensure the value is not changed until transaction is committed and also
// reuse the existing buffers and avoid allocations.
//
// This value is never written to, only read (should be safe for concurrent use)
//
// TODO:(gernest) In theory it must be a good approach however, benchmarks must
// support this. Is an extra copy better than extra allocation in this scenario ?
var base [4]byte

// saves mike to the badger key value store. For existing keys a is added to the
// existing value and the updated value is stored.
//
// a is stored as a big endian encoded byte slice. All keys and values rely on
// buffers provided by svc.ls to ensure they are visible and unmodified within the
// transaction that called this function. These buffers are not freed, it is the
// caller's responsibility to ensure svc.ls.Release() is called, which must be
// outside the transaction svc.txn (meaning after svc.txn.Commit())
func savePropertyKey(ctx context.Context, svc *saveContext, mike *bytes.Buffer, a uint32) error {
	svc.keys++
	key := mike.Bytes()
	b := svc.ls.Get()
	b.Write(base[:])
	value := b.Next(4)
	binary.BigEndian.PutUint32(value, a)
	return svc.txn.Set(key, value)
}
