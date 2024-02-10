package index

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"slices"
	"sync"
	"unsafe"

	"github.com/RoaringBitmap/roaring"
	"github.com/klauspost/compress/zstd"
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
	dataSize uint64
	r        ReaderAtSeeker
	meta     *v1.Metadata
	m        map[string]*FullColumn
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
	data, err := readChunk(r, meta.Offset, buf)
	if err != nil {
		return nil, err
	}
	raw := get()
	defer raw.Release()
	rawData, err := ZSTDDecompress(raw.get(int(meta.RawSize)), data)
	if err != nil {
		return nil, err
	}
	o := &FullColumn{
		name:    meta.Name,
		numRows: meta.NumRows,
		fst:     bytes.Clone(chuckFromRaw(rawData, meta.FstOffset)),
	}
	for _, bm := range meta.BitmapsOffset {
		b := new(roaring.Bitmap)
		err := b.UnmarshalBinary(chuckFromRaw(rawData, bm))
		if err != nil {
			return nil, err
		}
		o.bitmaps = append(o.bitmaps, b)
	}
	return o, nil
}

func chuckFromRaw(raw []byte, chunk *v1.Metadata_Chunk) []byte {
	return raw[chunk.Offset : chunk.Offset+chunk.Size]
}

func writeFull(w io.Writer, full Full, id string) error {
	b := new(bytes.Buffer)
	meta := &v1.Metadata{
		Id:  id,
		Min: full.Min(),
		Max: full.Max(),
	}
	var startOffset uint64
	err := full.Columns(func(column Column) (err error) {
		var col *v1.Metadata_Column
		col, startOffset, err = writeColumn(w, column, startOffset, b)
		if err == nil {
			meta.Columns = append(meta.Columns, col)
		}
		b.Reset()
		return err
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

func writeColumn(w io.Writer, column Column, startOffset uint64, b *bytes.Buffer) (meta *v1.Metadata_Column, offset uint64, err error) {
	meta = &v1.Metadata_Column{
		Name:    column.Name(),
		NumRows: column.NumRows(),
		Offset: &v1.Metadata_Chunk{
			Offset: startOffset,
		},
	}
	// fst is the first chunk
	n, err := w.Write(column.Fst())
	if err != nil {
		return nil, 0, err
	}
	meta.FstOffset = &v1.Metadata_Chunk{
		Offset: 0,
		Size:   uint64(n),
	}
	offset += uint64(n)

	column.Bitmaps(func(i int, b *roaring.Bitmap) error {
		data, err := b.MarshalBinary()
		if err != nil {
			return err
		}
		n, err := w.Write(data)
		if err != nil {
			return err
		}

		meta.BitmapsOffset = append(meta.BitmapsOffset, &v1.Metadata_Chunk{
			Offset: offset,
			Size:   uint64(n),
		})
		offset += uint64(n)
		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	meta.RawSize = uint32(b.Len())
	cb := get()
	defer cb.Release()
	o, err := ZSTDCompress(cb.dst(b.Len()), b.Bytes(), ZSTDCompressionLevel)
	if err != nil {
		return nil, 0, err
	}
	n, err = w.Write(o)
	if err != nil {
		return nil, 0, err
	}
	meta.Offset.Size = uint64(n)
	offset += uint64(n)
	return meta, startOffset + offset, nil
}

func readChunk(r ReaderAtSeeker, chunk *v1.Metadata_Chunk, b *compressBuf) ([]byte, error) {
	o := b.get(int(chunk.Size))
	n, err := r.ReadAt(o, int64(chunk.Offset))
	if err != nil {
		return nil, err
	}
	if n != int(chunk.Size) {
		return nil, fmt.Errorf("index: Too little data read want=%d got %d", chunk.Size, n)
	}
	return o, nil
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
