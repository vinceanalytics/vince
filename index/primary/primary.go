package primary

import (
	"cmp"
	"errors"
	"slices"
	"sort"
	"sync"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/cespare/xxhash/v2"
	"github.com/vinceanalytics/vince/db"
	v1 "github.com/vinceanalytics/vince/gen/go/staples/v1"
	"github.com/vinceanalytics/vince/index"
	"github.com/vinceanalytics/vince/logger"
	"google.golang.org/protobuf/proto"
)

var Key = []byte("index")

type PrimaryIndex struct {
	mu       sync.RWMutex
	resource string
	base     *v1.PrimaryIndex
	granules map[string][]*v1.Granule
	sites    map[string][]*roaring64.Bitmap
	db       db.Storage
}

func LoadPrimaryIndex(o *v1.PrimaryIndex, storage db.Storage) *PrimaryIndex {
	p := &PrimaryIndex{
		db:       storage,
		base:     &v1.PrimaryIndex{Resources: make(map[string]*v1.PrimaryIndex_Resource)},
		granules: make(map[string][]*v1.Granule),
		sites:    make(map[string][]*roaring64.Bitmap),
	}
	for r, x := range o.Resources {
		gs := make([]*v1.Granule, 0, len(x.Granules))
		for _, g := range x.Granules {
			gs = append(gs, g)
		}
		sort.Slice(gs, func(i, j int) bool {
			return gs[i].Min < gs[j].Min
		})
		sites := make([]*roaring64.Bitmap, len(gs))
		for i := range gs {
			b := new(roaring64.Bitmap)
			err := b.UnmarshalBinary(gs[i].Sites)
			if err != nil {
				logger.Fail("Failed to Unmarshal sites bitmap", "err", err)
			}
			sites[i] = b
		}
		p.granules[r] = gs
		p.sites[r] = sites
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
		sites:    make(map[string][]*roaring64.Bitmap),
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
	b := new(roaring64.Bitmap)
	err := b.UnmarshalBinary(granule.Sites)
	if err != nil {
		logger.Fail("Failed to Unmarshal sites bitmap", "err", err)
	}
	p.sites[resource] = append(p.sites[resource], b)
	data, _ := proto.Marshal(p.base)
	p.mu.Unlock()
	err = p.db.Set(Key, data, 0)
	if err != nil {
		panic("failed saving primary index " + err.Error())
	}
}

func (p *PrimaryIndex) FindGranules(resource string, start, end int64, siteId string) (o []string) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	gs := p.granules[resource]
	if len(gs) == 0 {
		return []string{}
	}

	from, _ := slices.BinarySearchFunc(gs, start, func(g *v1.Granule, i int64) int {
		return cmp.Compare(g.Min, gs[i].Min)
	})
	if from == len(gs) {
		return []string{}
	}
	to, _ := slices.BinarySearchFunc(gs, end, func(g *v1.Granule, i int64) int {
		return cmp.Compare(g.Min, gs[i].Min)
	})

	sites := p.sites[resource]
	h := new(xxhash.Digest)
	h.WriteString(siteId)
	domain := h.Sum64()

	if from == to {
		g := gs[from]
		if !index.Accept(g.Min, g.Max, start, end) {
			return
		}
		if !sites[from].Contains(domain) {
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
		if !sites[i].Contains(domain) {
			continue
		}
		o = append(o, g.Id)
	}
	return
}
