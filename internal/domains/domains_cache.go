package domains

import (
	"hash/crc32"
	"sync"

	"github.com/RoaringBitmap/roaring/v2"
	"github.com/gernest/len64/internal/kv"
)

type Cache struct {
	b  roaring.Bitmap
	mu sync.RWMutex
}

func New(db kv.KeyValue) (*Cache, error) {
	c := new(Cache)
	h := crc32.NewIEEE()
	return c, kv.Domains(db, func(domain string) {
		h.Reset()
		h.Write([]byte(domain))
		c.b.Add(h.Sum32())
	})
}

func (c *Cache) Contains(domain string) (ok bool) {
	h := crc32.NewIEEE()
	h.Write([]byte(domain))
	c.mu.RLock()
	ok = c.b.Contains(h.Sum32())
	c.mu.RUnlock()
	return
}

func (c *Cache) Add(domain string) {
	h := crc32.NewIEEE()
	h.Write([]byte(domain))
	c.mu.Lock()
	c.b.Add(h.Sum32())
	c.mu.Unlock()
}

func (c *Cache) Remove(domain string) {
	h := crc32.NewIEEE()
	h.Write([]byte(domain))
	c.mu.Lock()
	c.b.Remove(h.Sum32())
	c.mu.Unlock()
}
