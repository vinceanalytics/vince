package index

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"sync"
	"unsafe"

	"github.com/RoaringBitmap/roaring"
	"github.com/klauspost/compress/zstd"
	"github.com/vinceanalytics/vince/buffers"
	"github.com/vinceanalytics/vince/filters"
	v1 "github.com/vinceanalytics/vince/gen/go/staples/v1"
	"github.com/vinceanalytics/vince/logger"
	"google.golang.org/protobuf/proto"
)

const (
	footerSize = 8
)

type ReaderAtSeeker interface {
	io.ReaderAt
	io.Seeker
}

type FileIndex struct {
	r    ReaderAtSeeker
	meta *v1.Metadata
	m    map[string]*FullColumn
}

func NewFileIndex(r ReaderAtSeeker) (*FileIndex, error) {
	meta, err := readMetadata(r)
	if err != nil {
		return nil, err
	}
	_, err = r.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}
	return &FileIndex{
		r:    r,
		meta: meta,
		m:    make(map[string]*FullColumn),
	}, nil
}

var _ Full = (*FileIndex)(nil)

var baseFileIdxSize = uint64(unsafe.Sizeof(FileIndex{}))

func (f *FileIndex) Columns(_ func(column Column) error) error {
	logger.Fail("FileIndex does not support Columns indexing")
	return nil
}

func (f *FileIndex) Min() uint64 {
	return f.meta.Min
}

func (f *FileIndex) CanIndex() bool {
	return false
}

func (f *FileIndex) Max() uint64 {
	return f.meta.Min
}

func (f *FileIndex) Size() (n uint64) {
	n = baseFileIdxSize
	n += uint64(proto.Size(f.meta))
	n += uint64(len(f.m))
	for k, v := range f.m {
		n += uint64(len(k))
		n += v.Size()
	}
	return
}

func (f *FileIndex) Match(b *roaring.Bitmap, m []*filters.CompiledFilter) {
	for _, x := range m {
		v, err := f.get(x.Column)
		if err != nil {
			logger.Fail(err.Error())
		}
		b.And(v.Match(x))
	}
}

func (f *FileIndex) get(name string) (*FullColumn, error) {
	if c, ok := f.m[name]; ok {
		return c, nil
	}
	for _, c := range f.meta.GetColumns() {
		if c.Name == name {
			col, err := readColumn(f.r, c)
			if err != nil {
				return nil, err
			}
			f.m[name] = col
			return col, nil
		}
	}
	return nil, fmt.Errorf("Missing Index Column %v", name)
}

func readColumn(r ReaderAtSeeker, meta *v1.Metadata_Column) (*FullColumn, error) {
	buf := get()
	defer buf.Release()

	compress := get()
	defer compress.Release()

	raw := buf.get(int(meta.Size))
	n, err := r.ReadAt(raw, int64(meta.Offset))
	if err != nil {
		return nil, err
	}
	if n != int(meta.Size) {
		return nil, fmt.Errorf("index: Too little data read want=%d got %d", meta.Size, n)
	}
	data, err := ZSTDDecompress(compress.get(int(meta.RawSize)), raw)
	if err != nil {
		return nil, err
	}
	o := &FullColumn{
		name:    meta.Name,
		numRows: meta.NumRows,
		fst:     bytes.Clone(data[:meta.FstOffset]),
	}
	data = data[meta.FstOffset:]
	rd := bytes.NewReader(data)
	for {
		b := new(roaring.Bitmap)
		_, err := b.ReadFrom(rd)
		if err != nil {
			if strings.Contains(err.Error(), "EOF") {
				break
			}
			return nil, err
		}
		o.bitmaps = append(o.bitmaps, b)
	}
	return o, nil
}

func chuckFromRaw(raw []byte, chunk *v1.Metadata_Chunk) []byte {
	return raw[chunk.Offset : chunk.Offset+chunk.Size]
}

func WriteFull(w io.Writer, full Full, id string) error {
	b := buffers.Bytes()
	defer b.Release()
	compress := get()
	defer compress.Release()

	meta := &v1.Metadata{
		Id:  id,
		Min: full.Min(),
		Max: full.Max(),
	}
	var startOffset uint64
	err := full.Columns(func(column Column) (err error) {
		data, offset, err := writeColumn(column, b)
		if err != nil {
			return err
		}
		out, err := ZSTDCompress(
			compress.get(ZSTDCompressBound(len(data))),
			data, ZSTDCompressionLevel,
		)
		if err != nil {
			return err
		}
		n, err := w.Write(out)
		if err != nil {
			return err
		}
		meta.Columns = append(meta.Columns, &v1.Metadata_Column{
			Name:      column.Name(),
			NumRows:   column.NumRows(),
			FstOffset: uint32(offset),
			Offset:    startOffset,
			Size:      uint32(n),
			RawSize:   uint32(len(data)),
		})
		startOffset += uint64(n)
		return
	})
	if err != nil {
		return err
	}
	data, err := proto.Marshal(meta)
	if err != nil {
		return err
	}
	footer := make([]byte, 8)
	binary.BigEndian.PutUint32(footer[:], uint32(len(data)))
	_, err = w.Write(data)
	if err != nil {
		return err
	}
	_, err = w.Write(footer)
	return err
}

func writeColumn(column Column, w *buffers.BytesBuffer) (data []byte, offset int, err error) {
	w.Reset()
	// fst is the first chunk
	offset, err = w.Write(column.Fst())
	if err != nil {
		return nil, 0, err
	}
	column.Bitmaps(func(i int, b *roaring.Bitmap) error {
		_, err := b.WriteTo(w)
		return err
	})
	if err != nil {
		return nil, 0, err
	}
	data = w.Bytes()
	return
}

func readMetadata(r ReaderAtSeeker) (*v1.Metadata, error) {
	offset, err := r.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}
	if offset <= footerSize {
		return nil, fmt.Errorf("index: file too small %v", err)
	}
	buf := make([]byte, footerSize)
	n, err := r.ReadAt(buf, offset-int64(footerSize))
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("index: could not read footer: %w", err)
	}
	if n != len(buf) {
		return nil, fmt.Errorf("index: could not read %d bytes from end of file", len(buf))
	}
	size := int64(binary.BigEndian.Uint32(buf[:4]))
	if size < 0 || size+int64(footerSize) > offset {
		return nil, errors.New("index: file is smaller than indicated metadata size")
	}
	buf = make([]byte, size)
	if _, err := r.ReadAt(buf, offset-int64(footerSize)-size); err != nil {
		return nil, fmt.Errorf("index: could not read footer: %w", err)
	}
	var meta v1.Metadata
	err = proto.Unmarshal(buf, &meta)
	if err != nil {
		return nil, err
	}
	return &meta, nil
}

const ZSTDCompressionLevel = 3 // 1, 3, 9

var (
	decoder *zstd.Decoder
	encoder *zstd.Encoder

	encOnce, decOnce sync.Once
	compressPool     = &sync.Pool{New: func() any { return new(compressBuf) }}
)

type compressBuf struct {
	b []byte
}

func get() *compressBuf {
	return compressPool.Get().(*compressBuf)
}

func (c *compressBuf) Release() {
	clear(c.b)
	c.b = c.b[:0]
	compressPool.Put(c)
}

func (c *compressBuf) get(n int) []byte {
	c.b = slices.Grow(c.b, n)[:n]
	return c.b
}

func (c *compressBuf) dst(src int) []byte {
	return c.get(ZSTDCompressBound(src))
}

// ZSTDDecompress decompresses a block using ZSTD algorithm.
func ZSTDDecompress(dst, src []byte) ([]byte, error) {
	decOnce.Do(func() {
		var err error
		decoder, err = zstd.NewReader(nil)
		if err != nil {
			logger.Fail("index: failed creating zstd decompressor", "err", err)
		}
	})
	return decoder.DecodeAll(src, dst[:0])
}

// ZSTDCompress compresses a block using ZSTD algorithm.
func ZSTDCompress(dst, src []byte, compressionLevel int) ([]byte, error) {
	encOnce.Do(func() {
		var err error
		level := zstd.EncoderLevelFromZstd(compressionLevel)
		encoder, err = zstd.NewWriter(nil, zstd.WithEncoderLevel(level))
		if err != nil {
			logger.Fail("index: failed creating zstd compressor", "err", err)
		}
	})
	return encoder.EncodeAll(src, dst[:0]), nil
}

// ZSTDCompressBound returns the worst case size needed for a destination buffer.
// Klauspost ZSTD library does not provide any API for Compression Bound. This
// calculation is based on the DataDog ZSTD library.
// See https://pkg.go.dev/github.com/DataDog/zstd#CompressBound
func ZSTDCompressBound(srcSize int) int {
	lowLimit := 128 << 10 // 128 kB
	var margin int
	if srcSize < lowLimit {
		margin = (lowLimit - srcSize) >> 11
	}
	return srcSize + (srcSize >> 8) + margin
}
