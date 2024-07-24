package len64

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/RoaringBitmap/roaring/v2/roaring64"
	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/ipc"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/cespare/xxhash/v2"
	"github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/vfs"
	"google.golang.org/protobuf/proto"
)

type Store[T proto.Message] struct {
	db *pebble.DB
}

func New[T proto.Message](path string, mem bool) (*Store[T], error) {
	o := &pebble.Options{
		Merger: &pebble.Merger{
			Name:  "Lent64",
			Merge: merge,
		},
	}
	if mem {
		o.FS = vfs.NewMem()
	}
	db, err := pebble.Open(path, o)
	if err != nil {
		return nil, err
	}
	return &Store[T]{db: db}, nil
}

func (db *Store[T]) Batch() (*Batch[T], error) {
	return newBatch[T](db.db)
}

func (db *Store[T]) Compact() error {
	return db.db.Compact([]byte{shardPrefix}, seqKey, true)
}

func (db *Store[T]) Close() error {
	return db.db.Close()
}

// To avoid reading same bsi multiple times we cache them per shard.
type View struct {
	bsi   map[string]*roaring64.BSI
	snap  *pebble.Snapshot
	shard uint64
}

func (v *View) Reset(shard uint64) {
	clear(v.bsi)
	v.shard = shard
}

func (v *View) Get(name string) (*roaring64.BSI, error) {
	if b, ok := v.bsi[name]; ok {
		return b, nil
	}
	b, err := ReadBSI(v.snap, v.shard, name)
	if err != nil {
		return nil, err
	}
	v.bsi[name] = b
	return b, nil
}

func (db *Store[T]) Select(
	start, end time.Time, domain string,
	filter Filter,
	project []string,
) (Projection, error) {
	match := map[string][]*roaring64.BSI{}
	err := db.view(start, end, domain, func(db *View, foundSet *roaring64.Bitmap) error {
		f, err := filter.Apply(db, foundSet)
		if err != nil {
			return err
		}
		if f.IsEmpty() {
			return nil
		}
		for i := range project {
			b, err := db.Get(project[i])
			if err != nil {
				return err
			}
			newBSI := b.NewBSIRetainSet(f)
			match[project[i]] = append(match[project[i]], newBSI)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	o := make(Projection)
	for k, v := range match {
		if len(v) == 1 {
			// fast path to avoid extra work
			o[k] = v[0]
			continue
		}
		v[0].ParOr(parallel(), v[1:]...)
		o[k] = v[0]
	}
	return o, nil
}

func (db *Store[T]) view(start, end time.Time, domain string, f func(db *View, foundSet *roaring64.Bitmap) error) error {
	snap := db.db.NewSnapshot()
	defer snap.Close()
	shards := roaring64.New()
	from := start.UnixMilli()
	to := end.UnixMilli()
	err := ReadTimeRange(snap, uint64(from), uint64(to), shards)
	if err != nil {
		return err
	}

	hash := xxhash.New()
	hash.WriteString("domain")
	hash.Write(sep)
	hash.WriteString(domain)

	sum := hash.Sum64()

	it := shards.Iterator()
	view := &View{
		bsi:  make(map[string]*roaring64.BSI),
		snap: snap,
	}
	for it.HasNext() {
		shard := it.Next()
		view.Reset(shard)
		site, err := view.Get(domain)
		if err != nil {
			return fmt.Errorf("reading domain bsi%w", err)
		}
		match := site.CompareValue(parallel(), roaring64.EQ, int64(sum), 0, nil)
		if match.IsEmpty() {
			continue
		}
		ts, err := view.Get("timestamp")
		if err != nil {
			return fmt.Errorf("reading timestamp field%w", err)
		}
		match = ts.CompareValue(parallel(), roaring64.RANGE, from, to, match)
		if match.IsEmpty() {
			continue
		}
		err = f(view, match)
		if err != nil {
			return err
		}
	}
	return nil
}

func merge(key, value []byte) (pebble.ValueMerger, error) {
	if bytes.HasPrefix(key, dataPrefix) {
		return newData(key, value)
	}
	if bytes.HasPrefix(key, bsiPrefix) {
		return newBsi(key, value)
	}
	return noop(value), nil
}

type noop []byte

func (a noop) MergeNewer(value []byte) error {
	return nil
}

func (a noop) MergeOlder(value []byte) error {
	return nil
}

func (a noop) Finish(includesBase bool) ([]byte, io.Closer, error) {
	return []byte(a), nil, nil
}

type data struct {
	name string
	b    [2]arrow.Array
}

func newData(key, value []byte) (*data, error) {
	a, err := arrayFrom(value)
	if err != nil {
		return nil, err
	}
	x := &data{name: string(key[10:])}
	x.b[0] = a
	return x, nil
}

func (a *data) MergeNewer(value []byte) error {
	n, err := arrayFrom(value)
	if err != nil {
		return err
	}
	a.b[1] = n
	return a.apply()
}

func (a *data) MergeOlder(value []byte) error {
	n, err := arrayFrom(value)
	if err != nil {
		return err
	}
	a.b[1] = a.b[0]
	a.b[0] = n
	return a.apply()
}

func (a *data) Finish(includesBase bool) ([]byte, io.Closer, error) {
	defer func() {
		a.b[0].Release()
		a.b[0] = nil
	}()
	var b bytes.Buffer
	w := ipc.NewWriter(&b)
	err := w.Write(array.NewRecord(
		arrow.NewSchema(
			[]arrow.Field{
				{Name: a.name, Type: a.b[0].DataType()},
			},
			nil,
		),
		[]arrow.Array{a.b[0]},
		int64(a.b[0].Len()),
	))
	if err != nil {
		return nil, nil, err
	}
	return b.Bytes(), nil, nil
}

func (a *data) apply() error {
	n, err := array.Concatenate(a.b[:], memory.DefaultAllocator)
	if err != nil {
		return err
	}
	a.b[0].Release()
	a.b[1].Release()
	clear(a.b[:])
	a.b[0] = n
	return nil
}

type bsi struct {
	name string
	b    [2]*roaring64.BSI
}

func newBsi(key, value []byte) (*bsi, error) {
	a, err := bsiFrom(value)
	if err != nil {
		return nil, err
	}
	x := &bsi{name: string(key[10:])}
	x.b[0] = a
	return x, nil
}

func (a *bsi) MergeNewer(value []byte) error {
	n, err := bsiFrom(value)
	if err != nil {
		return err
	}
	a.b[1] = n
	return a.apply()
}

func (a *bsi) MergeOlder(value []byte) error {
	n, err := bsiFrom(value)
	if err != nil {
		return err
	}
	a.b[1] = a.b[0]
	a.b[0] = n
	return a.apply()
}

func (a *bsi) Finish(includesBase bool) ([]byte, io.Closer, error) {
	defer func() {
		clear(a.b[:])
	}()
	var b bytes.Buffer
	a.b[0].RunOptimize()
	_, err := a.b[0].WriteTo(&b)
	if err != nil {
		return nil, nil, err
	}
	return b.Bytes(), nil, nil
}

func (a *bsi) apply() error {
	a.b[0].ParOr(0, a.b[1])
	clear(a.b[:])
	a.b[1] = nil
	return nil
}
