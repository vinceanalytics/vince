package seq

import (
	"errors"
	"os"
	"reflect"
	"sync"
	"unsafe"

	"github.com/vinceanalytics/vince/internal/util/file"
)

type Seq struct {
	mu   sync.RWMutex
	file *file.MmapFile
	data []uint64
}

func New(path string) (*Seq, error) {
	ma, err := file.OpenMmapFile(path, os.O_RDWR|os.O_CREATE, 8)
	if err != nil && !errors.Is(err, file.ErrCreatingNewFile) {
		return nil, err
	}
	s := new(Seq)
	s.file = ma
	s.data = bytesToUint64Slice(s.file.Data[:8])
	return s, nil
}

func (s *Seq) Close() error {
	return s.file.Close(-1)
}

func (s *Seq) Load() uint64 {
	s.mu.RLock()
	v := s.data[0]
	s.mu.RUnlock()
	return v
}

func (s *Seq) Next() uint64 {
	s.mu.Lock()
	s.data[0]++
	v := s.data[0]
	s.mu.Unlock()
	return v
}

func bytesToUint64Slice(b []byte) []uint64 {
	if len(b) == 0 {
		return nil
	}
	var u64s []uint64
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&u64s))
	hdr.Len = len(b) / 8
	hdr.Cap = hdr.Len
	hdr.Data = uintptr(unsafe.Pointer(&b[0]))
	return u64s
}
