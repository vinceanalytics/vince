package be

import (
	"context"

	deadlockpb "github.com/pingcap/kvproto/pkg/deadlock"
	"github.com/pingcap/kvproto/pkg/kvrpcpb"
	"github.com/pingcap/tidb/kv"
	"github.com/pingcap/tidb/parser/model"
	"github.com/tikv/client-go/v2/oracle"
	"github.com/tikv/client-go/v2/tikv"
)

var _ kv.Driver = (*Driver)(nil)
var _ kv.Storage = (*Store)(nil)
var _ kv.Transaction = (*Txn)(nil)

type Driver struct{}

func (d *Driver) Open(path string) (kv.Storage, error) {
	return nil, nil
}

type Store struct{}

func (s *Store) Begin(opts ...tikv.TxnOption) (kv.Transaction, error) {
	return nil, nil
}
func (s *Store) GetSnapshot(ver kv.Version) kv.Snapshot {
	return nil
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
	return ""
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
	kv.RetrieverMutator
	kv.AssertionProto
	kv.FairLockingController
}

func (txn *Txn) Size() int {
	return 0
}
func (txn *Txn) Mem() uint64 {
	return 0
}
func (txn *Txn) SetMemoryFootprintChangeHook(func(uint64)) {}

func (txn *Txn) Len() int {
	return 0
}
func (txn *Txn) Reset() {}
func (txn *Txn) Commit(context.Context) error {
	return nil
}
func (txn *Txn) Rollback() error {
	return nil
}
func (txn *Txn) String() string {
	return ""
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
	return false
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
	return nil
}
func (txn *Txn) SetVars(vars interface{}) {

}
func (txn *Txn) GetVars() interface{} {
	return nil
}
func (txn *Txn) BatchGet(ctx context.Context, keys []kv.Key) (map[string][]byte, error) {
	return nil, nil
}
func (txn *Txn) IsPessimistic() bool {
	return false
}
func (txn *Txn) CacheTableInfo(id int64, info *model.TableInfo) {

}

func (txn *Txn) GetTableInfo(id int64) *model.TableInfo {
	return nil
}

func (txn *Txn) SetDiskFullOpt(level kvrpcpb.DiskFullOpt) {
}
func (txn *Txn) ClearDiskFullOpt() {}

func (txn *Txn) GetMemDBCheckpoint() *tikv.MemDBCheckpoint {
	return nil
}

func (txn *Txn) RollbackMemDBToCheckpoint(*tikv.MemDBCheckpoint) {}

func (txn *Txn) UpdateMemBufferFlags(key []byte, flags ...kv.FlagsOp) {}
