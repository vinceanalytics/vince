package neo

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/apache/arrow/go/v14/parquet/file"
	"github.com/dgraph-io/badger/v4"
	"github.com/oklog/ulid/v2"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/entry"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/must"
	v1 "github.com/vinceanalytics/vince/proto/v1"
	"google.golang.org/protobuf/proto"
)

type ActiveBlock struct {
	dir   string
	hosts sync.Map
	ctx   sync.Map
}

func NewBlock(dir string) *ActiveBlock {
	return &ActiveBlock{dir: dir}
}

func (a *ActiveBlock) Save(ctx context.Context) {
	a.hosts.Range(func(key, value any) bool {
		// Push new events to a new buffer while processing current buffer in the
		// background
		a.hosts.Store(key.(string), entry.NewMulti())

		go a.save(ctx, key.(string), value.(*entry.MultiEntry), false)
		return true
	})
}

func (a *ActiveBlock) Shutdown(ctx context.Context) error {
	a.hosts.Range(func(key, value any) bool {
		a.save(ctx, key.(string), value.(*entry.MultiEntry), true)
		return true
	})
	a.ctx.Range(func(key, value any) bool {
		value.(*writeContext).commit(ctx, key.(string))
		return true
	})
	return nil
}

func (a *ActiveBlock) Close() error {
	a.Shutdown(context.Background())
	return nil
}

func (a *ActiveBlock) WriteEntry(e *entry.Entry) {
	h, ok := a.hosts.Load(e.Domain)
	if !ok {
		m := entry.NewMulti()
		m.Append(e)
		e.Release()
		a.hosts.Store(e.Domain, m)
		return
	}
	h.(*entry.MultiEntry).Append(e)
	e.Release()
}

type writeContext struct {
	id string
	w  *file.Writer
	i  *v1.Block_Index
}

func (w *writeContext) commit(ctx context.Context, domain string) {
	must.One(w.w.Close())("closing parquet file writer ")
	b := must.Must(proto.Marshal(w.i))("marshalling index")
	key := keys.BlockIndex(domain, w.id)
	db.Get(ctx).Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), b)
	})
}

func (a *ActiveBlock) wctx(domain string) *writeContext {
	df, ok := a.ctx.Load(domain)
	if !ok {
		id := ulid.Make().String()
		w := &writeContext{
			id: id,
			w: entry.NewFileWriter(must.Must(os.Create(filepath.Join(a.dir, id)))(
				"failed creating active stats file", "file", id,
			)),
			i: &v1.Block_Index{
				Groups: make(map[int32]*v1.Block_Index_Bitmap),
			},
		}
		a.ctx.Store(domain, w)
		return w
	}
	return df.(*writeContext)
}

func (a *ActiveBlock) save(ctx context.Context, domain string, m *entry.MultiEntry, shutdown bool) {
	defer m.Release()
	w := a.wctx(domain)
	r := roaring64.New()
	m.Write(w.w, r)
	w.i.Groups[int32(w.w.NumRowGroups())-1] = &v1.Block_Index_Bitmap{
		Bitmap: must.Must(r.MarshalBinary())(
			"failed encoding binary",
		),
	}
	if w.w.NumRows() >= (1<<20) || shutdown {
		// Keep blocks sizes under 1 mb
		a.ctx.Delete(domain)
		w.commit(ctx, domain)
	}
}
