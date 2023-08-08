package neo

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/apache/arrow/go/v13/parquet/file"
	"github.com/dgraph-io/badger/v4"
	"github.com/oklog/ulid/v2"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/pkg/entry"
	v1 "github.com/vinceanalytics/vince/proto/v1"
	"google.golang.org/protobuf/proto"
)

const (
	blockPrefix    = "block/"
	metadataPrefix = "meta/"
	indexPrefix    = "index/"
)

type ActiveBlock struct {
	mu    sync.Mutex
	dir   string
	hosts map[string]*entry.MultiEntry
	ctx   sync.Map
}

func NewBlock(dir string) *ActiveBlock {
	return &ActiveBlock{
		hosts: make(map[string]*entry.MultiEntry),
	}
}

func (a *ActiveBlock) Save(ctx context.Context) {
	a.mu.Lock()
	if len(a.hosts) == 0 {
		a.mu.Unlock()
		return
	}
	for k, v := range a.hosts {
		go a.save(ctx, k, v, false)
		delete(a.hosts, k)
	}
	a.mu.Unlock()
}

func (a *ActiveBlock) Shutdown(ctx context.Context) {
	a.mu.Lock()
	if len(a.hosts) == 0 {
		a.mu.Unlock()
		return
	}
	for k, v := range a.hosts {
		go a.save(ctx, k, v, true)
		delete(a.hosts, k)
	}
	a.mu.Unlock()
}

func (a *ActiveBlock) Close() error {
	a.Shutdown(context.Background())
	return nil
}

func (a *ActiveBlock) WriteEntry(e *entry.Entry) {
	a.mu.Lock()
	h, ok := a.hosts[e.Domain]
	if !ok {
		h = entry.NewMulti()
		a.hosts[e.Domain] = h
	}
	a.mu.Unlock()
	h.Append(e)
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
	key := (&v1.Block_Key{
		Kind:   v1.Block_Key_INDEX,
		Domain: domain,
		Uid:    w.id,
	}).Badger()
	db.Get(ctx).Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), b)
	})
}

func (a *ActiveBlock) save(ctx context.Context, domain string, m *entry.MultiEntry, shutdown bool) {
	defer m.Release()
	df, ok := a.ctx.Load(domain)
	if !ok {
		if shutdown {
			return
		}
		id := ulid.Make().String()
		x := must.Must(os.Create(filepath.Join(a.dir, id)))()
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
	if w.i.Min == 0 {
		w.i.Min = m.Timestamp.First()
	}
	w.i.Max = m.Timestamp.Last()
	w.i.Groups[int32(w.w.NumRowGroups())-1] = &v1.Block_Index_Bitmap{
		Min:    m.Timestamp.First(),
		Max:    m.Timestamp.Last(),
		Bitmap: must.Must(r.MarshalBinary())(),
	}
	size := must.Must(w.f.Stat())().Size()
	if size >= (1<<20) || shutdown {
		// Keep blocks sizes under 1 mb
		a.ctx.Delete(domain)
		w.commit(ctx, domain)
	}
}
