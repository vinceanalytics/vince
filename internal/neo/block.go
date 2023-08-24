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
		a.hosts.Delete(key.(string))
		go a.save(ctx, key.(string), value.(*entry.MultiEntry), false)
		return true
	})
}

func (a *ActiveBlock) Shutdown(ctx context.Context) {
	a.hosts.Range(func(key, value any) bool {
		a.hosts.Delete(key.(string))
		go a.save(ctx, key.(string), value.(*entry.MultiEntry), true)
		return true
	})
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
	f  *os.File
	w  *file.Writer
	i  *v1.Block_Index
}

func (w *writeContext) commit(ctx context.Context, domain string) {
	must.One(w.w.Close())("closing parquet file writer for ", domain)
	must.One(w.f.Close())("closing parquet file  for ", domain)
	b := must.Must(proto.Marshal(w.i))("marshalling index")
	key := keys.BlockIndex(domain, w.id)
	db.Get(ctx).Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), b)
	})
}

func (a *ActiveBlock) save(ctx context.Context, domain string, m *entry.MultiEntry, shutdown bool) {
	defer m.Release()
	df, ok := a.ctx.Load(domain)
	if !ok {
		id := ulid.Make().String()
		x := must.Must(os.Create(filepath.Join(a.dir, id)))(
			"failed creating active stats file", "file", id,
		)
		w := &writeContext{
			id: id,
			f:  x,
			w:  entry.NewFileWriter(x),
			i: &v1.Block_Index{
				Groups: make(map[int32]*v1.Block_Index_Bitmap),
			},
		}
		a.ctx.Store(domain, w)
		df = w
	}
	w := df.(*writeContext)
	r := roaring64.New()
	m.Write(w.w, r)
	lo, hi := m.Boundary()
	if w.i.Min == 0 {
		// w.i.Min tracks block wise minimum value. Timestamps are assumed to always
		// increase. Any non zero minimum value encountered is the one which will
		// cover this whole block
		w.i.Min = lo
	}
	w.i.Max = max(w.i.Max, hi)

	w.i.Groups[int32(w.w.NumRowGroups())-1] = &v1.Block_Index_Bitmap{
		Min: lo,
		Max: hi,
		Bitmap: must.Must(r.MarshalBinary())(
			"failed encoding binary",
		),
	}
	size := must.Must(w.f.Stat())(
		"failed getting stats for active file", "name", w.f.Name(),
	).Size()
	if size >= (1<<20) || shutdown {
		// Keep blocks sizes under 1 mb
		a.ctx.Delete(domain)
		w.commit(ctx, domain)
	}
}
