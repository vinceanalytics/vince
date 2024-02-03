package primary

import (
	"errors"
	"slices"
	"sort"
	"sync"

	"github.com/vinceanalytics/staples/staples/db"
	v1 "github.com/vinceanalytics/staples/staples/gen/go/staples/v1"
	"google.golang.org/protobuf/proto"
)

var Key = []byte("index")

type PrimaryIndex struct {
	mu       sync.RWMutex
	resource string
	base     *v1.PrimaryIndex
	stamps   map[string][]int64
	ids      map[string][]string
	db       db.Storage
}

func LoadPrimaryIndex(o *v1.PrimaryIndex, storage db.Storage) *PrimaryIndex {
	p := &PrimaryIndex{
		db:     storage,
		base:   &v1.PrimaryIndex{Resources: make(map[string]*v1.PrimaryIndex_Resource)},
		stamps: make(map[string][]int64),
		ids:    make(map[string][]string),
	}
	for r, x := range o.Resources {
		gs := make([]*v1.Granule, 0, len(x.Granules))
		for _, g := range x.Granules {
			gs = append(gs, g)
		}
		sort.Slice(gs, func(i, j int) bool {
			return gs[i].Min < gs[j].Min
		})
		ts := make([]int64, len(gs))
		ids := make([]string, len(gs))
		for i := range gs {
			ts[i] = gs[i].Min
			ids[i] = gs[i].Id
		}
		p.stamps[r] = ts
		p.ids[r] = ids
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
		db:     store,
		base:   &v1.PrimaryIndex{Resources: make(map[string]*v1.PrimaryIndex_Resource)},
		stamps: make(map[string][]int64),
		ids:    make(map[string][]string),
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
	p.stamps[resource] = append(p.stamps[resource], granule.Min)
	p.ids[resource] = append(p.ids[resource], granule.Id)
	data, _ := proto.Marshal(p.base)
	p.mu.Unlock()
	err := p.db.Set(Key, data, 0)
	if err != nil {
		panic("failed saving primary index " + err.Error())
	}
}

func (p *PrimaryIndex) FindGranules(resource string, start, end int64) []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	ts := p.stamps[resource]
	if len(ts) == 0 {
		return []string{}
	}
	ids := p.ids[resource]

	from, _ := slices.BinarySearch(ts, start)
	if from == len(ts) {
		return []string{}
	}
	to, _ := slices.BinarySearch(ts, end)
	if to == len(ts) {
		return slices.Clone(ids[from:])
	}
	return slices.Clone(ids[from:to])
}
