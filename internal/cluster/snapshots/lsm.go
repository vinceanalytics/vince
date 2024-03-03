package snapshots

import (
	"io"
	"os"

	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/ipc"
	"github.com/apache/arrow/go/v15/arrow/memory"
	"github.com/hashicorp/raft"
	"github.com/vinceanalytics/vince/internal/lsm"
)

type Persist interface {
	Persist(f lsm.CompactCallback)
}

type Arrow struct {
	File *os.File
	Tree Persist
	Mem  memory.Allocator
}

var _ raft.FSMSnapshot = (*Arrow)(nil)

func (a *Arrow) Persist(sink raft.SnapshotSink) (err error) {
	return a.Backup(sink)
}

func (a *Arrow) Backup(sink io.Writer) (err error) {
	w, werr := ipc.NewFileWriter(
		a.File, ipc.WithAllocator(a.Mem),
		ipc.WithZstd(),
	)
	if werr != nil {
		return werr
	}
	a.Tree.Persist(func(r arrow.Record) bool {
		if err != nil {
			return false
		}
		err = w.Write(r)
		return true
	})
	if err != nil {
		return
	}
	err = w.Close()
	if err != nil {
		return
	}
	_, err = a.File.WriteTo(sink)
	return
}

func (a *Arrow) Release() {
	a.File.Close()
	os.Remove(a.File.Name())
}

type ArrowRestore struct {
	Mem  memory.Allocator
	File *os.File
}

var _ lsm.RecordSource = (*ArrowRestore)(nil)

func (a *ArrowRestore) Record(f func(arrow.Record) error) error {
	rd, err := ipc.NewFileReader(a.File, ipc.WithAllocator(a.Mem))
	if err != nil {
		return err
	}
	for {
		r, err := rd.Read()
		if err != nil {
			return err
		}
		err = f(r)
		if err != nil {
			return err
		}
	}
}
