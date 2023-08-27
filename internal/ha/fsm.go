package ha

import (
	"io"
	"runtime"

	"github.com/dgraph-io/badger/v4"
	"github.com/hashicorp/raft"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/proto/bpb"
	"google.golang.org/protobuf/proto"
)

type FSM struct {
	db db.Provider
}

var _ raft.FSM = (*FSM)(nil)

type fsmSnap struct {
	db db.Provider
}

func (f *FSM) Apply(l *raft.Log) interface{} {
	if l.Type == raft.LogCommand {
		return f.db.With(func(d *badger.DB) error {
			var ls bpb.KVList
			err := proto.Unmarshal(l.Data, &ls)
			if err != nil {
				return err
			}
			ld := d.NewKVLoader(1)
			defer ld.Finish()
			x := bpb.To(&ls)
			for _, v := range x.Kv {
				err = ld.Set(v)
				if err != nil {
					return err
				}
			}
			return nil
		})
	}
	return nil
}

func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	return &fsmSnap{db: f.db}, nil
}

func (f *FSM) Restore(r io.ReadCloser) error {
	return f.db.With(func(db *badger.DB) error {
		return db.Load(r, runtime.NumCPU())
	})
}

var _ raft.FSMSnapshot = (*fsmSnap)(nil)

func (f *fsmSnap) Persist(w raft.SnapshotSink) error {
	return f.db.With(func(db *badger.DB) error {
		_, err := db.Backup(w, 0)
		if err != nil {
			w.Cancel()
			return err
		}
		return w.Close()
	})
}

func (f *fsmSnap) Release() {}
