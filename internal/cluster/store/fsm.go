package store

import (
	"io"

	"github.com/hashicorp/raft"
)

type FSM struct {
	store *Store
}

// NewFSM returns a new FSM.
func NewFSM(s *Store) *FSM {
	return &FSM{store: s}
}

var _ raft.FSM = (*FSM)(nil)

func (f *FSM) Apply(l *raft.Log) interface{} {
	return f.store.fsmApply(l)
}

func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	return f.store.fsmSnapshot()
}

func (f *FSM) Restore(w io.ReadCloser) error {
	return f.store.fsmRestore(w)
}
