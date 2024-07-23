package Len64

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"regexp"

	"github.com/RoaringBitmap/roaring/v2/roaring64"
	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/ipc"
	"github.com/cespare/xxhash/v2"
	"github.com/cockroachdb/pebble"
)

var (
	shardPrefix     = byte(0x0)
	timeRangePrefix = byte(0x1)
	dataPrefix      = []byte{shardPrefix, 0x0}
	bsiPrefix       = []byte{shardPrefix, 0x1}
	trKeyPrefix     = []byte{shardPrefix, 0x2}
	sep             = []byte{'='}
)

func WriteRecord(b *pebble.Batch, shard uint64, r arrow.Record) error {
	var buf bytes.Buffer
	key := make([]byte, 1<<10)
	copy(key, dataPrefix)
	binary.BigEndian.PutUint64(key[2:], shard)

	for i := 0; i < int(r.NumCols()); i++ {
		column := r.Column(i)
		name := r.ColumnName(i)
		buf.Reset()

		w := ipc.NewWriter(&buf)
		err := w.Write(array.NewRecord(
			arrow.NewSchema(
				[]arrow.Field{
					{Name: name, Type: column.DataType()},
				},
				nil,
			),
			[]arrow.Array{column},
			int64(column.Len()),
		))
		if err != nil {
			return fmt.Errorf("writing column %d:%s%w", shard, name, err)
		}
		err = w.Close()
		if err != nil {
			return fmt.Errorf("close writing column %d:%s%w", shard, name, err)
		}
		key = append(key[:10], []byte(name)...)

		err = b.Merge(key, buf.Bytes(), nil)
		if err != nil {
			return fmt.Errorf("merge writing column %d:%s%w", shard, name, err)
		}
	}
	return nil
}

func ReadArray(db *pebble.DB, shard uint64, name string) (arrow.Array, error) {
	key := make([]byte, 1<<10)
	copy(key, dataPrefix)
	binary.BigEndian.PutUint64(key[2:], shard)
	key = append(key[:10], []byte(name)...)

	value, done, err := db.Get(key)
	if err != nil {
		return nil, err
	}
	defer done.Close()

	r, err := ipc.NewReader(bytes.NewReader(value), ipc.WithDelayReadSchema(true))
	if err != nil {
		return nil, err
	}
	defer r.Release()
	r.Next()
	record := r.Record()
	// we only have single field records
	a := record.Column(0)
	a.Retain() // avoid this bing released
	return a, nil
}

func WriteBSI(b *pebble.Batch, shard uint64, m map[string]*roaring64.BSI) error {
	key := make([]byte, 1<<10)
	copy(key, bsiPrefix)
	binary.BigEndian.PutUint64(key[2:], shard)

	var buf bytes.Buffer

	for name, v := range m {
		key = append(key[:10], []byte(name)...)
		buf.Reset()
		v.RunOptimize()
		_, err := v.WriteTo(&buf)
		if err != nil {
			return fmt.Errorf("writing bsi %d:%s%w", shard, name, err)
		}

		err = b.Merge(key, buf.Bytes(), nil)
		if err != nil {
			return fmt.Errorf("merge bsi %d:%s%w", shard, name, err)
		}
	}
	return nil
}

func ReadBSI(db *pebble.DB, shard uint64, name string) (*roaring64.BSI, error) {
	key := make([]byte, 1<<10)
	copy(key, bsiPrefix)
	binary.BigEndian.PutUint64(key[2:], shard)
	key = append(key[:10], []byte(name)...)

	value, done, err := db.Get(key)
	if err != nil {
		return nil, err
	}
	defer done.Close()
	r := roaring64.NewDefaultBSI()
	_, err = r.ReadFrom(bytes.NewReader(value))
	if err != nil {
		return nil, err
	}
	return r, nil
}

func WriteString(b *pebble.Batch, shard uint64, m map[uint64]string) error {
	key := make([]byte, 1<<10)
	copy(key, trKeyPrefix)
	binary.BigEndian.PutUint64(key[2:], shard)

	for _, v := range m {
		key = append(key[:10], []byte(v)...)
		err := b.Set(key, []byte{}, nil)
		if err != nil {
			return fmt.Errorf("write string key %d:%s%w", shard, v, err)
		}
	}
	return nil
}

func Search(db *pebble.DB, shard uint64, k, v string) (uint64, bool) {
	var b bytes.Buffer
	b.Write(trKeyPrefix)
	var x [8]byte
	binary.BigEndian.PutUint64(x[:], shard)
	b.Write(x[:])
	b.WriteString(k)
	b.WriteByte('=')
	b.WriteString(v)
	full := b.Bytes()
	_, done, err := db.Get(full)
	if err != nil {
		return 0, false
	}
	done.Close()

	return xxhash.Sum64(full[10:]), true
}

func SearchRegex(db *pebble.DB, shard uint64, k, v string) ([]uint64, error) {
	var b bytes.Buffer
	b.Write(trKeyPrefix)
	var x [8]byte
	binary.BigEndian.PutUint64(x[:], shard)
	b.Write(x[:])
	b.WriteString(k)
	b.WriteByte('=')

	re, err := regexp.Compile(v)
	if err != nil {
		return nil, err
	}
	prefix := b.Bytes()

	it, err := db.NewIter(nil)
	if err != nil {
		return nil, err
	}
	result := make([]uint64, 0, 4)
	h := new(xxhash.Digest)
	for it.SeekGE(prefix); bytes.HasPrefix(it.Key(), prefix); it.Next() {
		full := it.Key()
		str := full[10:]
		_, value, _ := bytes.Cut(str, sep)
		if re.Match(value) {
			h.Reset()
			h.Write(str)
			result = append(result, h.Sum64())
		}
	}
	return result, nil
}

func WriteTimeRange(b *pebble.Batch, shard uint64, min, max uint64) error {
	key := make([]byte, 1+8+8)
	key[0] = timeRangePrefix
	binary.BigEndian.PutUint64(key[1:], min)
	binary.BigEndian.PutUint64(key[1+8:], shard)

	err := b.Set(key, []byte{}, nil)
	if err != nil {
		return fmt.Errorf("write time range  %d:%d%w", shard, min, err)
	}
	binary.BigEndian.PutUint64(key[1:], max)
	err = b.Set(key, []byte{}, nil)
	if err != nil {
		return fmt.Errorf("write time range  %d:%d%w", shard, max, err)
	}
	return nil
}

func ReadTimeRange(db *pebble.DB, start, end uint64, b *roaring64.Bitmap) error {
	key := make([]byte, 1+8+8)
	key[0] = timeRangePrefix
	binary.BigEndian.PutUint64(key[1:], start)
	binary.BigEndian.PutUint64(key[1+8:], 0)

	from := bytes.Clone(key)
	binary.BigEndian.PutUint64(key[1:], end)
	binary.BigEndian.PutUint64(key[1+8:], math.MaxUint64)

	it, err := db.NewIter(&pebble.IterOptions{
		LowerBound: from,
		UpperBound: key,
	})
	if err != nil {
		return err
	}
	defer it.Close()
	for it.First(); it.Valid(); it.Next() {
		b.Add(
			binary.BigEndian.Uint64(it.Key()[1+8:]),
		)
	}
	return nil
}
