package be

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/dgraph-io/badger/v4"
	deadlockpb "github.com/pingcap/kvproto/pkg/deadlock"
	"github.com/pingcap/kvproto/pkg/kvrpcpb"
	"github.com/pingcap/tidb/kv"
	"github.com/pingcap/tidb/parser/model"
	"github.com/tikv/client-go/v2/oracle"
	"github.com/tikv/client-go/v2/tikv"
	"github.com/vinceanalytics/vince/internal/must"
)

type Store struct {
	db *badger.DB
}

var _ kv.Storage = (*Store)(nil)

func (s *Store) Begin(opts ...tikv.TxnOption) (kv.Transaction, error) {
	return &Txn{db: s.db, ts: uint64(time.Now().UnixMilli())}, nil
}

func (s *Store) GetSnapshot(ver kv.Version) kv.Snapshot {
	return &Txn{db: s.db}
}

func (s *Store) GetClient() kv.Client {
	return nil
}

func (s *Store) GetMPPClient() kv.MPPClient {
	return nil
}

func (s *Store) Close() error {
	return nil
}

func (s *Store) UUID() string {
	return "vince"
}

func (s *Store) CurrentVersion(txnScope string) (kv.Version, error) {
	return kv.Version{}, nil
}

func (s *Store) GetOracle() oracle.Oracle {
	return nil
}
func (s *Store) SupportDeleteRange() (supported bool) {
	return false
}
func (s *Store) Name() string {
	return ""
}

func (s *Store) Describe() string {
	return ""
}

func (s *Store) ShowStatus(ctx context.Context, key string) (interface{}, error) {
	return nil, nil
}
func (s *Store) GetMemCache() kv.MemManager {
	return nil
}
func (s *Store) GetMinSafeTS(txnScope string) uint64 {
	return 0
}
func (s *Store) GetLockWaits() ([]*deadlockpb.WaitForEntry, error) {
	return nil, nil
}
func (s *Store) GetCodec() tikv.Codec {
	return nil
}

type Txn struct {
	ts   uint64
	vars interface{}
	db   *badger.DB
	kv.RetrieverMutator
	kv.AssertionProto
	kv.FairLockingController
}

var _ kv.Transaction = (*Txn)(nil)
var _ kv.Snapshot = (*Txn)(nil)

func (txn *Txn) Size() int {
	l, v := txn.db.Size()
	return int(l + v)
}

func (txn *Txn) Mem() uint64 {
	return 0
}
func (txn *Txn) SetMemoryFootprintChangeHook(func(uint64)) {}

func (txn *Txn) Len() (n int) {
	for _, t := range txn.db.Tables() {
		n += int(t.KeyCount)
	}
	return
}

func (txn *Txn) Reset() {}

func (txn *Txn) Commit(context.Context) error {
	return nil
}

func (txn *Txn) Rollback() error {
	return nil
}

func (txn *Txn) String() string {
	return strconv.FormatUint(txn.ts, 10)
}

func (txn *Txn) LockKeys(ctx context.Context, lockCtx *kv.LockCtx, keys ...kv.Key) error {
	return nil
}

func (txn *Txn) LockKeysFunc(ctx context.Context, lockCtx *kv.LockCtx, fn func(), keys ...kv.Key) error {
	return nil
}
func (txn *Txn) SetOption(opt int, val interface{}) {}
func (txn *Txn) GetOption(opt int) interface{} {
	return nil
}

func (txn *Txn) IsReadOnly() bool {
	return true
}

func (txn *Txn) StartTS() uint64 {
	return 0
}
func (txn *Txn) Valid() bool {
	return false
}
func (txn *Txn) GetMemBuffer() kv.MemBuffer {
	return nil
}

func (txn *Txn) GetSnapshot() kv.Snapshot {
	return txn
}

func (txn *Txn) SetVars(vars interface{}) {
	txn.vars = vars
}
func (txn *Txn) GetVars() interface{} {
	return txn.vars
}

func (txn *Txn) BatchGet(ctx context.Context, keys []kv.Key) (m map[string][]byte, err error) {
	err = txn.db.View(func(txn *badger.Txn) error {
		m = make(map[string][]byte)
		for i := range keys {
			it, err := txn.Get(keys[i])
			if err != nil {
				return err
			}
			m[string(keys[i])] = must.Must(it.ValueCopy(nil))()
		}
		return nil
	})
	return
}
func (txn *Txn) IsPessimistic() bool {
	return false
}
func (txn *Txn) CacheTableInfo(id int64, info *model.TableInfo) {

}

func (txn *Txn) GetTableInfo(id int64) *model.TableInfo {
	return nil
}

func (txn *Txn) SetDiskFullOpt(level kvrpcpb.DiskFullOpt) {}

func (txn *Txn) ClearDiskFullOpt() {}

func (txn *Txn) GetMemDBCheckpoint() *tikv.MemDBCheckpoint {
	return nil
}

func (txn *Txn) RollbackMemDBToCheckpoint(*tikv.MemDBCheckpoint) {}

func (txn *Txn) UpdateMemBufferFlags(key []byte, flags ...kv.FlagsOp) {}

func (txn *Txn) Get(ctx context.Context, k kv.Key) (o []byte, err error) {
	err = txn.db.View(func(txn *badger.Txn) error {
		it, err := txn.Get(k)
		if err != nil {
			return err
		}
		o = must.Must(it.ValueCopy(nil))()
		return nil
	})
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil, kv.ErrNotExist
		}
	}
	return
}

func (txn *Txn) Iter(k kv.Key, upperBound kv.Key) (kv.Iterator, error) {
	x := txn.db.NewTransaction(false)
	o := badger.DefaultIteratorOptions
	o.Prefix = k
	it := x.NewIterator(o)
	it.Rewind()
	return &iterator{
		hi: upperBound,
		i:  it, txn: x,
	}, nil
}

func (txn *Txn) IterReverse(k kv.Key, upperBound kv.Key) (kv.Iterator, error) {
	x := txn.db.NewTransaction(false)
	o := badger.DefaultIteratorOptions
	o.Prefix = k
	o.Reverse = true
	it := x.NewIterator(o)

	it.Rewind()
	return &iterator{
		hi: upperBound,
		i:  it, txn: x,
	}, nil
}

type iterator struct {
	hi  kv.Key
	txn *badger.Txn
	i   *badger.Iterator
}

var _ kv.Iterator = (*iterator)(nil)

func (i *iterator) Valid() (ok bool) {
	ok = i.i.Valid()
	if ok && i.hi != nil {
		key := kv.Key(i.i.Item().Key())
		ok = key.Cmp(i.hi) == -1
	}
	return
}

func (i *iterator) Key() kv.Key {
	return i.i.Item().KeyCopy(nil)
}

func (i *iterator) Value() []byte {
	return must.Must(i.i.Item().ValueCopy(nil))()
}

func (i *iterator) Next() error {
	i.i.Next()
	return nil
}

func (i *iterator) Close() {
	i.i.Close()
	i.txn.Discard()
}
