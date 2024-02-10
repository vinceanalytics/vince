package index

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"slices"
	"sync"

	"github.com/RoaringBitmap/roaring"
	"github.com/klauspost/compress/zstd"
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

func writeFull(w io.Writer, full Full) error {
	chunk := new(chunkWriter)
	b := new(bytes.Buffer)
	meta := &v1.Metadata{
		Min: full.Min(),
		Max: full.Max(),
	}
	var startOffset uint64
	err := full.Columns(func(column Column) (err error) {
		var col *v1.Metadata_Column
		col, startOffset, err = writeColumn(w, column, startOffset, chunk, b)
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

func writeColumn(w io.Writer, column Column,
	startOffset uint64,
	chunk *chunkWriter,
	b *bytes.Buffer,
) (meta *v1.Metadata_Column, offset uint64, err error) {
	chunk.reset(b)
	meta = &v1.Metadata_Column{
		Name:    column.Name(),
		NumRows: column.NumRows(),
		Offset:  startOffset,
	}
	meta.FstOffset, err = chunk.write(column.Fst())
	if err != nil {
		return nil, 0, err
	}
	column.Bitmaps(func(i int, b *roaring.Bitmap) error {
		data, err := b.MarshalBinary()
		if err != nil {
			return err
		}
		offset, err := chunk.write(data)
		if err != nil {
			return err
		}
		meta.BitmapsOffset = append(meta.BitmapsOffset, offset)
		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	meta.RawSize = uint32(b.Len())
	cb := get()
	defer cb.Release()
	chunk.reset(w)
	o, err := ZSTDCompress(cb.dst(b.Len()), b.Bytes(), ZSTDCompressionLevel)
	if err != nil {
		return nil, 0, err
	}
	_, err = chunk.write(o)
	if err != nil {
		return nil, 0, err
	}
	return meta, startOffset + uint64(chunk.offset), nil
}

type chunkWriter struct {
	w       io.Writer
	scratch [8]byte
	offset  uint32
}

func (s *chunkWriter) reset(w io.Writer) {
	s.w = w
	s.offset = 0
}

func (s *chunkWriter) write(b []byte) (offset uint32, err error) {
	_, err = s.w.Write(binary.BigEndian.AppendUint32(s.scratch[:4], uint32(len(b))))
	if err != nil {
		return 0, err
	}
	offset = s.offset
	_, err = s.w.Write(b)
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

func (c *compressBuf) dst(src int) []byte {
	c.b = slices.Grow(c.b, ZSTDCompressBound(src))
	return c.b
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
