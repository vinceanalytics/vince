package timeseries

import (
	"bytes"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/klauspost/compress/zstd"
)

var compressPool = &sync.Pool{
	New: func() any {
		// we really care about the size of data that goes into badger.
		var c compressor
		e, _ := zstd.NewWriter(&c.b, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
		c.e = e
		return &c
	},
}

type compressor struct {
	e *zstd.Encoder
	b bytes.Buffer
}

func (c *compressor) Release() {
	c.b.Reset()
	c.e.Reset(&c.b)
	compressPool.Put(c)
}

// only compress blocks exceeding 1 kb
const compressThreshold = 1 << 10

func (c *compressor) Compress(b []byte) []byte {
	c.b.Reset()
	c.e.Reset(&c.b)
	c.e.Write(b)
	return c.b.Bytes()
}

func getCompressor() *compressor {
	return compressPool.Get().(*compressor)
}

func (c *compressor) Write(txn *badger.Txn, key, value []byte, ttl time.Duration) error {
	if len(value) < compressThreshold {
		return txn.SetEntry(badger.NewEntry(key, value).WithTTL(ttl))
	}
	return txn.SetEntry(badger.NewEntry(key, c.Compress(value)).WithMeta(1).WithTTL(ttl))
}

type decompressor struct {
	d *zstd.Decoder
	b bytes.Buffer
}

func getDecompressor() *decompressor {
	return decompressorPool.Get().(*decompressor)
}
func (d *decompressor) Release() {
	d.b.Reset()
	d.d.Reset(nil)
	decompressorPool.Put(d)
}

func (d *decompressor) Decode(b []byte) []byte {
	d.b.Reset()
	d.d.Reset(bytes.NewReader(b))
	d.d.WriteTo(&d.b)
	return d.b.Bytes()
}

func (d *decompressor) Read(it *badger.Item, f func(val []byte) error) error {
	if it.UserMeta() != 1 {
		// not compressed value
		return it.Value(f)
	}
	return it.Value(func(val []byte) error {
		return f(d.Decode(val))
	})
}

var decompressorPool = &sync.Pool{
	New: func() any {
		var d decompressor
		z, _ := zstd.NewReader(nil)
		d.d = z
		return &d
	},
}
