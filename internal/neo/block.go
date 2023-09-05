package neo

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"github.com/oklog/ulid/v2"
	"github.com/parquet-go/parquet-go"
	"github.com/thanos-io/objstore"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/entry"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/internal/px"
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
	a.get(e.Domain).append(a.Context, e)
}

type writeContext struct {
	id, domain string
	mu         sync.Mutex
	name       string
	w          *parquet.GenericWriter[*entry.Entry]
	log        *slog.Logger
	o          objstore.Bucket
	scratch    [1]*entry.Entry
}

func (w *writeContext) append(ctx context.Context, e *entry.Entry) {
	w.mu.Lock()
	w.scratch[0] = e
	w.w.Write(w.scratch[:])
	w.mu.Unlock()
}

func (w *writeContext) Save(ctx context.Context) {
	w.commit(ctx)
}

func (w *writeContext) commit(ctx context.Context) {
	w.log.Info("committing active block")
	must.One(w.w.Close())("closing parquet file writer ")
	// We make sure we commit metadata after we have successfully uploaded the
	// block to the object store. This avoids having metadata about blocks that are
	// not in the permanent storage
	w.upload(ctx)
	w.index(ctx)
	w.cleanup(ctx)
}

func (w *writeContext) Release() {
	w.id = ""
	w.domain = ""
	w.name = ""
	w.w.Reset(nil)
	w.log = nil
	w.o = nil
	w.scratch[0] = nil
	writePool.Put(w)

}

var writePool = &sync.Pool{New: func() any { return &writeContext{} }}

func (a *Ingest) get(domain string) *writeContext {
	df, ok := a.ctx.Load(domain)
	if !ok {
		id := ulid.Make().String()
		file := must.Must(os.CreateTemp("", "vince"))("failed creating temporary file for block write")
		w := writePool.Get().(*writeContext)
		if w.w == nil {
			w.w = parquet.NewGenericWriter[*entry.Entry](file)
		}
		w.domain = domain
		w.id = id
		w.name = file.Name()
		w.o = a.os
		w.log = a.log.With(
			slog.String("block", id),
			slog.String("domain", domain),
			slog.String("temp_file", file.Name()),
		)
		a.ctx.Store(domain, w)
		return w
	}
	return df.(*writeContext)
}

func (w *writeContext) upload(ctx context.Context) {
	w.log.Info("uploading block to permanent storage")
	f := must.Must(os.Open(w.name))("failed opening block file")
	must.One(w.o.Upload(ctx, w.id, f))("failed uploading block to permanent storage")
	f.Close()
}

func (w *writeContext) index(ctx context.Context) {
	w.log.Info("indexing block")
	f := must.Must(os.Open(w.name))("failed opening block file")
	db.Get(ctx).Txn(true, func(txn db.Txn) error {
		key := keys.BlockIndex(w.domain, w.id)
		defer key.Release()
		idx := must.Must(IndexBlockFile(f))("failed to index the block file")
		return txn.Set(key.Bytes(), px.Encode(idx))
	})
	f.Close()
}

func (w *writeContext) cleanup(ctx context.Context) {
	w.log.Info("removing temporary file")
	must.One(os.Remove(w.name))("failed removing uploaded block file")
}
