package timeseries

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/system"
	"github.com/vinceanalytics/vince/pkg/log"
)

type Sum struct {
	Visitors,
	Views,
	Events,
	Visits,
	BounceRate uint16
	VisitDuration time.Duration
}

func DropSite(ctx context.Context, uid, sid uint64) {
	start := core.Now(ctx)
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
	start := core.Now(ctx).UTC().Truncate(time.Millisecond)
	startMs := uint64(start.UnixMilli())

	db := GetMike(ctx)
	meta := newMetaKey()

	svc := saveContext{
		slice: newSlice(),
		ls:    newTxnBufferList(),
	}

	defer func() {
		log.Get().Debug().Int(
			"__size__", len(b.segments.Timestamp),
		).Int("__keys__", svc.keys).
			Msg("saved stats")
		system.EntriesPerBufferSave.Observe(float64(len(b.segments.Timestamp)))
		system.KeysPerBufferSave.Observe(float64(svc.keys))
		system.SaveDuration.Observe(time.Since(start).Seconds())
		b.Release()
		meta.Release()
		svc.ls.Release()
		svc.slice.release()
	}()

	// Buffer.id has the same encoding as the first 16 bytes of meta. We just copy that
	// there is no need to re encode user id and site id.
	copy(meta[:], b.id[:])
	var tsBytes [6]byte
	setTs(tsBytes[:], startMs)
	svc.txn = db.NewTransactionAt(startMs, true)
	err := b.Build(ctx, func(p Property, key string, sum *Sum) error {
		return saveProperty(ctx, &svc, tsBytes[:], meta.prop(p), key, sum)
	})
	if err != nil {
		log.Get().Err(err).Msg("failed to save ts buffer")
		svc.txn.Discard()
	} else {
		err = svc.txn.CommitAt(startMs, nil)
		if err != nil {
			// Failing to commit transaction is fatal
			log.Get().Fatal().Err(err).Msg("failed to commit transaction")
		}
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
		savePropertyKey(ctx, svc, m.metric(VisitDurations).key(ts, text, svc.ls), uint16(a.VisitDuration.Milliseconds())),
	)
}

type saveContext struct {
	txn   *badger.Txn
	ls    *txnBufferList
	slice *slice
	keys  int
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

func savePropertyKey(ctx context.Context, svc *saveContext, key *bytes.Buffer, a uint16) error {
	if a == 0 {
		return nil
	}
	svc.keys++
	k := key.Bytes()
	println(">", DebugKey(k), a)
	return svc.txn.Set(k, svc.slice.u16(a))
}

type slice struct {
	d   []byte
	pos int
}

var slicePool = &sync.Pool{
	New: func() any {
		return &slice{
			d: make([]byte, 0, 1<<10),
		}
	},
}

func newSlice() *slice {
	return slicePool.Get().(*slice)
}

func (s *slice) u16(v uint16) []byte {
	o := s.grow()
	binary.BigEndian.PutUint16(o, v)
	return o
}

func (s *slice) grow() []byte {
	if cap(s.d) < s.pos+2 {
		o := make([]byte, cap(s.d)*2)
		copy(o, s.d)
		s.d = o
	}
	s.d = s.d[:s.pos+2]
	o := s.pos
	s.pos += 2
	return s.d[o:]
}

func (s *slice) release() {
	s.pos = 0
	s.d = s.d[:0]
	slicePool.Put(s)
}
