package snapshots

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/hashicorp/raft"
)

type Badger struct {
	DB *badger.DB
}

func NewBadger(db *badger.DB) *Badger {
	return &Badger{DB: db}
}

var _ raft.FSMSnapshot = (*Badger)(nil)

func (b *Badger) Persist(sink raft.SnapshotSink) error {
	_, err := b.DB.Backup(sink, 0)
	return err
}

func (Badger) Release() {}
