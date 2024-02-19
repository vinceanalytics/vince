package primary

import (
	"cmp"
	"errors"
	"slices"
	"sort"
	"sync"

	v1 "github.com/vinceanalytics/vince/gen/go/staples/v1"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/index"
	"google.golang.org/protobuf/proto"
)

var Key = []byte("index")

type PrimaryIndex struct {
	mu       sync.RWMutex
	resource string
	base     *v1.PrimaryIndex
	granules map[string][]*v1.Granule
	db       db.Storage
}

func LoadPrimaryIndex(o *v1.PrimaryIndex, storage db.Storage) *PrimaryIndex {
	p := &PrimaryIndex{
		db:       storage,
		base:     &v1.PrimaryIndex{Resources: make(map[string]*v1.PrimaryIndex_Resource)},
		granules: make(map[string][]*v1.Granule),
	}
	for r, x := range o.Resources {
		gs := make([]*v1.Granule, 0, len(x.Granules))
		for _, g := range x.Granules {
			gs = append(gs, g)
		}
		sort.Slice(gs, func(i, j int) bool {
			return gs[i].Min < gs[j].Min
		})
		p.granules[r] = gs
	}
	return p
}

func NewPrimary(store db.Storage) (idx *PrimaryIndex, err error) {
	err = store.Get(Key, func(b []byte) error {
		var o v1.PrimaryIndex
		err := proto.Unmarshal(b, &o)
		if err != nil {
			return err
		}
		idx = LoadPrimaryIndex(&o, store)
		return nil
	})
	if !errors.Is(err, db.ErrKeyNotFound) {
		return
	}
	return &PrimaryIndex{
		db:       store,
		base:     &v1.PrimaryIndex{Resources: make(map[string]*v1.PrimaryIndex_Resource)},
		granules: make(map[string][]*v1.Granule),
	}, nil
}

func (p *PrimaryIndex) Add(resource string, granule *v1.Granule) {
	p.mu.Lock()
	r, ok := p.base.Resources[resource]
	if !ok {
		r = &v1.PrimaryIndex_Resource{
			Name:     resource,
			Granules: make(map[string]*v1.Granule),
		}
		p.base.Resources[resource] = r
	}
	r.Granules[granule.Id] = granule
	p.granules[resource] = append(p.granules[resource], granule)
	data, _ := proto.Marshal(p.base)
	p.mu.Unlock()
	err := p.db.Set(Key, data, 0)
	if err != nil {
		panic("failed saving primary index " + err.Error())
	}
}

func (p *PrimaryIndex) FindGranules(resource string, start, end int64) (o []string) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	gs := p.granules[resource]
	if len(gs) == 0 {
		return []string{}
	}

	from, _ := slices.BinarySearchFunc(gs, start, func(g *v1.Granule, i int64) int {
		return cmp.Compare(g.Min, i)
	})

	if from == len(gs) {
		return []string{}
	}
	to, _ := slices.BinarySearchFunc(gs, end, func(g *v1.Granule, i int64) int {
		return cmp.Compare(g.Min, i)
	})

	if from == to {
		g := gs[from]
		if !index.Accept(g.Min, g.Max, start, end) {
			return
		}
		return []string{g.Id}
	}
	o = make([]string, 0, to-from)
	for i := from; i < to && i < len(gs); i++ {
		g := gs[i]
		if !index.Accept(g.Min, g.Max, start, end) {
			continue
		}
		o = append(o, g.Id)
	}
	return
}
