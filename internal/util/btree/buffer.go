package btree

import (
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/vinceanalytics/vince/internal/util/file"
)

type Source interface {
	AllocateOffset(size int) []byte
	Reset() []byte
	Release() error
}

type Slice struct {
	data []byte
}

var _ Source = (*Slice)(nil)

func (s *Slice) AllocateOffset(size int) []byte {
	if len(s.data)+size >= cap(s.data) {
		s.data = slices.Grow(s.data, size)
	}
	return s.data[:len(s.data)+size]
}

func (s *Slice) Reset() []byte {
	clear(s.data)
	s.data = s.data[:0]
	return s.data
}

func (s *Slice) Release() error { return nil }

type File struct {
	file   *file.MmapFile
	buf    []byte
	offset int64
	curSz  int
}

type FileTree = Tree[*File]

// covers up to 16368 key/value pairs and 16 pages.
const defaultFileSize = 256 << 10

func NewFileTree(path string) (*FileTree, error) {
	ma, err := file.OpenMmapFile(path, os.O_RDWR|os.O_CREATE, defaultFileSize)
	if err != nil && !errors.Is(err, file.ErrCreatingNewFile) {
		return nil, err
	}
	fs := &File{
		file:  ma,
		buf:   ma.Data,
		curSz: defaultFileSize,
	}
	t := new(FileTree)
	t.base = fs
	fs.offset = int64(len(fs.buf))
	t.data = fs.buf
	root := t.node(1)
	isInitialized := root.pageID() != 0
	if !isInitialized {
		t.nextPage = 1
		t.freePage = 0
		t.initRootNode()
	} else {
		t.reinit()
	}
	return t, nil
}

var _ Source = (*File)(nil)

func (f *File) Release() error {
	path := f.file.Fd.Name()
	if err := f.file.Close(f.offset); err != nil {
		return fmt.Errorf("%w while closing file: %s", err, path)
	}
	*f = File{}
	return nil
}

func (f *File) Reset() []byte {
	clear(f.buf)
	f.offset = 0
	return f.buf[:0]
}

func (f *File) AllocateOffset(n int) []byte {
	f.grow(n)
	f.offset += int64(n)
	return f.buf[:f.offset]
}

func (f *File) grow(n int) {
	if int(f.offset)+n < f.curSz {
		return
	}

	// Calculate new capacity.
	growBy := f.curSz + n
	// Don't allocate more than 1GB at a time.
	if growBy > 1<<30 {
		growBy = 1 << 30
	}
	// Allocate at least n, even if it exceeds the 1GB limit above.
	if n > growBy {
		growBy = n
	}
	f.curSz += growBy
	if err := f.file.Truncate(int64(f.curSz)); err != nil {
		err = fmt.Errorf(
			"%w while trying to truncate file: %s to size: %d", err, f.file.Fd.Name(), f.curSz)
		panic(err)
	}
	f.buf = f.file.Data
}
