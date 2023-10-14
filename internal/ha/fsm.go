package ha

import (
	"io"
	"runtime"

	"github.com/dgraph-io/badger/v4"
	"github.com/hashicorp/raft"
	raftv1 "github.com/vinceanalytics/proto/gen/go/vince/raft/v1"
	"github.com/vinceanalytics/vince/internal/db"
	"google.golang.org/protobuf/proto"
)

type fsm struct {
	base db.Provider
}

var _ raft.FSM = (*fsm)(nil)

func NewFSM(base db.Provider) raft.FSM {
	return &fsm{base: base}
}

func (f *fsm) Apply(l *raft.Log) interface{} {
	if l.Type == raft.LogCommand {
		return f.base.Txn(true, func(txn db.Txn) error {
			var e raftv1.RaftEntry
			err := proto.Unmarshal(l.Data, &e)
			if err != nil {
				return err
			}
			if e.Expires != nil {
				return txn.SetTTL(e.Key, e.Value, e.Expires.AsDuration())
			}
			return txn.Set(e.Key, e.Value)
		})
	}
	return nil
}

func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	return &fsmSnap{base: f.base}, nil
}

func (f *fsm) Restore(r io.ReadCloser) error {
	return f.base.With(func(db *badger.DB) error {
		return db.Load(r, runtime.NumCPU())
	})
}

type fsmSnap struct {
	base db.Provider
}

var _ raft.FSMSnapshot = (*fsmSnap)(nil)

func (f *fsmSnap) Persist(w raft.SnapshotSink) error {
	return f.base.With(func(db *badger.DB) error {
		_, err := db.Backup(w, 0)
		if err != nil {
			w.Cancel()
			return err
		}
		return w.Close()
	})
}

func (f *fsmSnap) Release() {}
