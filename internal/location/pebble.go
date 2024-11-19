package location

import (
	"embed"
	"io"
	"io/fs"
	"path/filepath"

	"github.com/cockroachdb/pebble/vfs"
)

//go:embed data
var allData embed.FS

type FS struct {
	vfs.FS
}

var _ vfs.FS = (*FS)(nil)

func (fs *FS) OpenDir(name string) (vfs.File, error) {
	f, err := allData.Open(name)
	if err != nil {
		return nil, err
	}
	return &file{base: f}, nil
}

func (*FS) GetDiskUsage(path string) (vfs.DiskUsage, error) {
	return vfs.DiskUsage{}, nil
}

func (fs *FS) Open(name string, opts ...vfs.OpenOption) (vfs.File, error) {
	f, err := allData.Open(name)
	if err != nil {
		return nil, err
	}
	x := &file{base: f}
	return x, nil
}

func (*FS) Lock(name string) (io.Closer, error) {
	return io.NopCloser(nil), nil
}

func (fs *FS) List(dir string) ([]string, error) {
	es, err := allData.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	o := make([]string, len(es))
	for i := range es {
		o[i] = es[i].Name()
	}
	return o, err
}

func (fs *FS) Stat(name string) (vfs.FileInfo, error) {
	f, err := allData.Open(name)
	if err != nil {
		return nil, err
	}
	s, _ := f.Stat()
	return &info{s}, err
}

func (fs *FS) PathBase(path string) string {
	return filepath.Base(path)
}

func (fs *FS) PathJoin(elem ...string) string {
	return filepath.Join(elem...)
}

func (fs *FS) PathDir(path string) string {
	return filepath.Dir(path)
}

type file struct {
	base fs.File
	vfs.File
}

var _ vfs.File = (*file)(nil)

func (f *file) Close() error {
	return f.base.Close()
}

func (f *file) ReadAt(p []byte, off int64) (n int, err error) {
	return f.base.(io.ReaderAt).ReadAt(p, off)
}

func (f *file) Read(p []byte) (int, error) {
	return f.base.Read(p)
}

func (f *file) Stat() (vfs.FileInfo, error) {
	s, _ := f.base.Stat()
	return &info{s}, nil
}

type info struct {
	fs.FileInfo
}

var _ vfs.FileInfo = (*info)(nil)

func (i *info) DeviceID() vfs.DeviceID {
	return vfs.DeviceID{}
}
