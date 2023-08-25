package neo

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/apache/arrow/go/v14/parquet/file"
	"github.com/dgraph-io/badger/v4"
	"github.com/oklog/ulid/v2"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/entry"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/must"
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
	w.m.Write(w.w)
	w.log.Debug("saved events to block",
		slog.Int("rows", w.w.NumRows()),
		slog.Int("groups", w.w.NumRowGroups()),
	)
}

func (w *writeContext) commit(ctx context.Context) {
	must.One(w.w.Close())("closing parquet file writer ")
	key := keys.BlockMeta(w.domain, w.id)
	var b bytes.Buffer
	must.Must(w.w.FileMetadata.WriteTo(&b, nil))("failed serializing block metadata")
	db.Get(ctx).Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), b.Bytes())
	})
	w.log.Debug("commit block",
		slog.Int("rows", w.w.NumRows()),
		slog.Int("groups", w.w.NumRowGroups()),
	)

	// we make sure we release the events buffer so we can reuse it
	w.m.Release()
	w.m = nil
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
