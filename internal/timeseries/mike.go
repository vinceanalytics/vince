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

type aggr struct {
	Visitors,
	Views,
	Events,
	Visits,
	BounceRate uint16
	VisitDuration time.Duration
}

func DropSite(ctx context.Context, uid, sid uint64) {
	start := core.Now(ctx)
	id := newMetaKey()
	id.uid(uid, sid)
	var wg sync.WaitGroup
	prefix := id[:metricOffset]
	sl := newSlice()
	defer func() {
		sl.release()
		id.Release()
		system.DropSiteDuration.Observe(core.Now(ctx).Sub(start).Seconds())
	}()
	wg.Add(2)
	ts := uint64(start.UnixMilli())
	go dropSiteTemporary(ctx, &wg, ts, sl.clone(prefix))
	go dropSitePermanent(ctx, &wg, ts, sl.clone(prefix))
	wg.Wait()
}

func dropSiteTemporary(ctx context.Context, wg *sync.WaitGroup, ts uint64, prefix []byte) {
	defer wg.Done()
	err := drop(GetMike(ctx), ts, prefix)
	if err != nil {
		log.Get().Err(err).Msg("failed to delete site data from temporary storage")
	}
}

func dropSitePermanent(ctx context.Context, wg *sync.WaitGroup, ts uint64, prefix []byte) {
	defer wg.Done()
	err := drop(Get(ctx), ts, prefix)
	if err != nil {
		log.Get().Err(err).Msg("failed to delete site data from permanent storage")
	}
}

func drop(db *badger.DB, ts uint64, prefix []byte) error {
	sl := newSlice()
	txn := db.NewTransactionAt(ts, false)
	delTxn := db.NewTransactionAt(ts, true)
	o := badger.DefaultIteratorOptions
	o.PrefetchValues = false
	o.Prefix = prefix
	it := txn.NewIterator(o)

	defer func() {
		it.Close()
		txn.Discard()
		delTxn.Discard()
		sl.release()
	}()
	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		clone := sl.clone(it.Item().Key())
		err := delTxn.Delete(clone)
		if err != nil {
			if errors.Is(err, badger.ErrTxnTooBig) {
				err = delTxn.CommitAt(ts, nil)
				if err != nil {
					return err
				}
				sl.reset()
				delTxn = db.NewTransactionAt(ts, true)
				err = delTxn.Delete(clone)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}
	return delTxn.CommitAt(ts, nil)
}

func Save(ctx context.Context, b *Buffer) {
	start := core.Now(ctx)
	startMs := uint64(start.UnixMilli())

	db := GetMike(ctx)
	meta := newMetaKey()

	svc := saveContext{
		slice: newSlice(),
		ls:    newTxnBufferList(),
	}

	// b.id has the same encoding as the first 16 bytes of meta. We just copy that
	// there is no need to re encode user id and site id.
	copy(meta[:], b.id[:])
	tsBytes := svc.slice.u64(uint64(start.Truncate(time.Hour).UnixMilli()))

	svc.txn = db.NewTransactionAt(startMs, true)
	err := b.build(ctx, func(p Property, key string, sum *aggr) error {
		return transaction(&svc, tsBytes, meta.prop(p), key, sum)
	})
	svc.commit(ctx, startMs, err)

	log.Get().Debug().Int(
		"__size__", len(b.segments.Timestamp),
	).Int("__keys__", svc.keys).
		Msg("saved stats")
	system.EntriesPerBufferSave.Observe(float64(len(b.segments.Timestamp)))
	system.KeysPerBufferSave.Observe(float64(svc.keys))
	system.SaveDuration.Observe(time.Since(start).Seconds())
	b.Release()
	meta.Release()
}

// save aggregate a to active transaction associated with svc. Th active
// transaction is never committed. It is up to the caller to manually commit the
// transaction.
//
// all keys are suffixed with encoded timestamp ts. This is specifically to
// provide context aware uniqueness for the keys. This is done to avoid managing
// extra versions of the same key. It comes with a cost of bigger keys, ts has
// no other use than this so it should be an active area when we need to optimize
// for space ( we can take first 4 bytes, since ts encodes significant bits
// first) saving 2 bytes per key.
//
// memory for keys and values is never released, values use svc.slice and keys
// use svc.ls for memory buffers. svc.ls svc.slice must be released after the
// transaction has been committed/discarded.
func transaction(
	svc *saveContext,
	ts []byte,
	m *Key, text string, a *aggr) error {
	return errors.Join(
		save(svc, m.metric(Visitors).key(ts, text, svc.ls), a.Visitors),
		save(svc, m.metric(Views).key(ts, text, svc.ls), a.Views),
		save(svc, m.metric(Events).key(ts, text, svc.ls), a.Events),
		save(svc, m.metric(Visits).key(ts, text, svc.ls), a.Visits),
		save(svc, m.metric(BounceRates).key(ts, text, svc.ls), a.BounceRate),
		save(svc, m.metric(VisitDurations).key(ts, text, svc.ls), uint16(a.VisitDuration)),
	)
}

type saveContext struct {
	txn   *badger.Txn
	ls    *txnBufferList
	slice *slice
	keys  int
}

func (svc *saveContext) commit(ctx context.Context, ms uint64, err error) {
	if err != nil {
		log.Get().Err(err).Msg("failed to save ts buffer")
		svc.txn.Discard()
	} else {
		err = svc.txn.CommitAt(ms, nil)
		if err != nil {
			// Failing to commit transaction is fatal. We don't want the app to run if we
			// can't commit changes to our permanent storage.
			log.Get().Fatal().Err(err).Msg("failed to commit transaction")
		}
	}
	svc.ls.release()
	svc.slice.release()
}

// keeps reference to keys. We need this to make sure we reuse ID and properly
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

func (ls *txnBufferList) release() {
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

func get() *bytes.Buffer {
	return smallBufferpool.Get().(*bytes.Buffer)
}

func put(b *bytes.Buffer) {
	b.Reset()
	smallBufferpool.Put(b)
}

// stores key and aggregate value a into the transaction associated with svc. If
// a is 0 this is a No Op call.
func save(svc *saveContext, key *bytes.Buffer, a uint16) error {
	if a == 0 {
		return nil
	}
	svc.keys++
	k := key.Bytes()
	// println(">", DebugKey(k), a)
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
	o := s.get(2)
	binary.BigEndian.PutUint16(o, v)
	return o
}

func (s *slice) u64(v uint64) []byte {
	o := s.get(8)
	binary.BigEndian.PutUint64(o, v)
	return o
}

func (s *slice) get(n int) []byte {
	if cap(s.d) < s.pos+n {
		o := make([]byte, cap(s.d)*2)
		copy(o, s.d)
		s.d = o
	}
	s.d = s.d[:s.pos+n]
	o := s.pos
	s.pos += n
	return s.d[o:]
}

func (s *slice) release() {
	s.reset()
	slicePool.Put(s)
}

func (s *slice) clone(b []byte) []byte {
	o := s.get(len(b))
	copy(o, b)
	return o
}
func (s *slice) reset() {
	s.pos = 0
	s.d = s.d[:0]
}
