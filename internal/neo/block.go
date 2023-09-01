package neo

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/apache/arrow/go/v14/parquet"
	"github.com/apache/arrow/go/v14/parquet/file"
	"github.com/apache/arrow/go/v14/parquet/metadata"
	"github.com/cespare/xxhash/v2"
	"github.com/oklog/ulid/v2"
	blocksv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/blocks/v1"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/entry"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/internal/px"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ActiveBlock struct {
	context.Context
	dir      string
	capacity int
	ctx      sync.Map
}

func NewBlock(ctx context.Context, dir string, capacity int) *ActiveBlock {
	return &ActiveBlock{dir: dir, Context: ctx, capacity: capacity}
}

func (a *ActiveBlock) Commit(domain string) {
	w, ok := a.ctx.LoadAndDelete(domain)
	if !ok {
		return
	}
	x := w.(*writeContext)
	x.save(a.Context)
	x.commit(a.Context)
}

func (a *ActiveBlock) Close() error {
	slog.Debug("closing active block")
	a.ctx.Range(func(key, value any) bool {
		a.ctx.Delete(key.(string))
		x := value.(*writeContext)
		x.save(a.Context)
		value.(*writeContext).commit(a.Context)
		return true
	})
	return nil
}

func (a *ActiveBlock) WriteEntry(e *entry.Entry) {
	a.wctx(e.Domain).append(a.Context, e)
}

type writeContext struct {
	id, domain string
	capacity   int
	mu         sync.Mutex
	m          *entry.MultiEntry
	w          *file.Writer
	log        *slog.Logger
	h          xxhash.Digest
	blooms     []*roaring64.Bitmap
}

var bitmapPool = &sync.Pool{New: func() any { return roaring64.New() }}

func bloom() *roaring64.Bitmap {
	return bitmapPool.Get().(*roaring64.Bitmap)
}

func releaseBloom(r *roaring64.Bitmap) {
	r.Clear()
	bitmapPool.Put(r)
}

func (w *writeContext) append(ctx context.Context, e *entry.Entry) {
	w.mu.Lock()
	w.m.Append(e)
	if w.m.Len() == w.capacity {
		w.save(ctx)
		w.m.Reset()
	}
	w.mu.Unlock()
	e.Release()
}

func (w *writeContext) save(ctx context.Context) {
	b := []byte{0}
	r := bloom()
	w.m.Write(w.w, func(c v1.Column, ba parquet.ByteArray) {
		if len(ba) == 0 {
			return
		}
		b[0] = byte(c)
		w.h.Reset()
		w.h.Write(b)
		w.h.Write(ba)
		r.CheckedAdd(w.h.Sum64())
	})
	w.blooms = append(w.blooms, r)
	w.log.Debug("saved events to block",
		slog.Int("rows", w.w.NumRows()),
		slog.Int("groups", w.w.NumRowGroups()),
	)
}

func (w *writeContext) commit(ctx context.Context) {
	must.One(w.w.Close())("closing parquet file writer ")
	idx := blocksv1.BlockIndex{
		RowGroupBitmap: make([][]byte, 0, len(w.blooms)),
		TimeRange:      make([]*blocksv1.BlockIndex_Range, 0, len(w.blooms)),
	}
	meta := w.w.FileMetadata.RowGroups
	for i, r := range w.blooms {
		idx.RowGroupBitmap = append(idx.RowGroupBitmap,
			must.Must(r.MarshalBinary())("failed serializing bitmap"),
		)
		g := meta[i].Columns[px.ColumnIndex(v1.Column_timestamp)].MetaData.Statistics
		idx.TimeRange = append(idx.TimeRange, &blocksv1.BlockIndex_Range{
			Min: timestamppb.New(
				time.UnixMilli(
					metadata.GetStatValue(parquet.Types.Int64, g.MinValue).(int64),
				),
			),
			Max: timestamppb.New(
				time.UnixMilli(
					metadata.GetStatValue(parquet.Types.Int64, g.MaxValue).(int64),
				),
			),
		})
	}
	index := must.Must(proto.Marshal(&idx))("failed serializing index")

	var b bytes.Buffer
	must.Must(w.w.FileMetadata.WriteTo(&b, nil))("failed serializing block metadata")
	err := db.Get(ctx).Txn(true, func(txn db.Txn) error {
		metaKey := keys.BlockMeta(w.domain, w.id)
		defer metaKey.Release()
		indexKey := keys.BlockIndex(w.domain, w.id)
		defer indexKey.Release()

		return errors.Join(
			txn.Set(metaKey.Bytes(), b.Bytes()),
			txn.Set(indexKey.Bytes(), index),
		)
	})
	must.One(err)("failed saving block metadata to meta storage")
	w.log.Debug("commit block",
		slog.Int("rows", w.w.NumRows()),
		slog.Int("groups", w.w.NumRowGroups()),
	)
	// we make sure we release the events buffer so we can reuse it
	w.m.Release()
	w.m = nil
	for i := range w.blooms {
		releaseBloom(w.blooms[i])
		w.blooms[i] = nil
	}
	w.blooms = nil
}

func (a *ActiveBlock) ReadBlock(id ulid.ULID, f func(parquet.ReaderAtSeeker)) {
	o, err := os.Open(filepath.Join(a.dir, id.String()))
	if err != nil {
		slog.Error("failed opening block", "id", id.String())
		return
	}
	f(o)
	o.Close()
}

func (a *ActiveBlock) wctx(domain string) *writeContext {
	df, ok := a.ctx.Load(domain)
	if !ok {
		id := ulid.Make().String()
		w := &writeContext{
			domain:   domain,
			capacity: a.capacity,
			id:       id,
			w: entry.NewFileWriter(must.Must(os.Create(filepath.Join(a.dir, id)))(
				"failed creating active stats file", "file", id,
			)),
			m: entry.NewMulti(),
			log: slog.Default().With(
				slog.String("block", id),
				slog.String("domain", domain),
			),
		}
		a.ctx.Store(domain, w)
		return w
	}
	return df.(*writeContext)
}
