package neo

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/apache/arrow/go/v14/parquet"
	"github.com/apache/arrow/go/v14/parquet/file"
	"github.com/apache/arrow/go/v14/parquet/metadata"
	"github.com/cespare/xxhash/v2"
	"github.com/oklog/ulid/v2"
	"github.com/thanos-io/objstore"
	blocksv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/blocks/v1"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/store/v1"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/entry"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/internal/px"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Ingest struct {
	context.Context
	capacity int
	ctx      sync.Map
	os       objstore.Bucket
	log      *slog.Logger
}

func NewIngest(ctx context.Context, o objstore.Bucket, capacity int) *Ingest {
	return &Ingest{os: o, Context: ctx, capacity: capacity, log: slog.Default().With("component", "ingest")}
}

func (a *Ingest) Close() error {
	a.log.Info("closing")
	a.Run(a.Context)
	return nil
}

// Implements worker.Job. Persists active blocks
func (a *Ingest) Run(ctx context.Context) {
	a.ctx.Range(func(key, value any) bool {
		a.ctx.Delete(key.(string))
		value.(*writeContext).Save(a.Context)
		return true
	})
}

func (a *Ingest) WriteEntry(e *entry.Entry) {
	a.wctx(e.Domain).append(a.Context, e)
}

type writeContext struct {
	id, domain string
	capacity   int
	mu         sync.Mutex
	m          *entry.MultiEntry
	f          *os.File
	w          *file.Writer
	log        *slog.Logger
	h          xxhash.Digest
	blooms     []*roaring64.Bitmap
	o          objstore.Bucket
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

func (w *writeContext) Save(ctx context.Context) {
	w.save(ctx)
	w.commit(ctx)
}

func (w *writeContext) save(ctx context.Context) {
	w.log.Info("saving buffered events")
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
	w.log.Info("saved events to block",
		slog.Int("rows", w.w.NumRows()),
		slog.Int("groups", w.w.NumRowGroups()),
	)
}

func (w *writeContext) commit(ctx context.Context) {
	w.log.Info("committing active block")
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

	// We make sure we commit metadata after we have successfully uploaded the
	// block to the object store. This avoids having metadata about blocks that are
	// not in the permanent storage
	w.upload(ctx)

	w.log.Info("saving block metadata")
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

func (a *Ingest) wctx(domain string) *writeContext {
	df, ok := a.ctx.Load(domain)
	if !ok {
		id := ulid.Make().String()
		file := must.Must(os.CreateTemp("", "vince"))("failed creating temporary file for block write")
		w := &writeContext{
			domain:   domain,
			capacity: a.capacity,
			id:       id,
			f:        file,
			o:        a.os,
			// calling w.Close will also close file, which we don't want. We need to keep
			// the file open until it is uploaded to the object store.
			//
			// We use noopClose to to make the close call on file a no op
			w: entry.NewFileWriter(file),
			m: entry.NewMulti(),
			log: a.log.With(
				slog.String("block", id),
				slog.String("domain", domain),
				slog.String("temp_file", file.Name()),
			),
		}
		a.ctx.Store(domain, w)
		return w
	}
	return df.(*writeContext)
}

func (w *writeContext) upload(ctx context.Context) {
	w.log.Info("uploading block to permanent storage")
	f := must.Must(os.Open(w.f.Name()))("failed opening block file")
	must.One(w.o.Upload(ctx, w.id, f))("failed uploading block to permanent storage")
	f.Close()
	must.One(os.Remove(f.Name()))("failed removing uploaded block file")
}
