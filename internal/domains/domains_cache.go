package domains

import (
	"hash/crc32"
	"sync"
	"time"

	"github.com/RoaringBitmap/roaring/v2"
	"github.com/gernest/len64/internal/kv"
	"go.etcd.io/bbolt"
	"golang.org/x/time/rate"
)

type Cache struct {
	b  roaring.Bitmap
	r  map[uint32]*rate.Limiter
	mu sync.RWMutex
}

func New(db *bbolt.DB) (*Cache, error) {
	c := &Cache{r: make(map[uint32]*rate.Limiter)}
	h := crc32.NewIEEE()
	limit := limit()
	return c, kv.Domains(db, func(domain string) {
		h.Reset()
		h.Write([]byte(domain))
		id := h.Sum32()
		c.b.Add(id)
		// We are strict with quota, zero burst given!
		c.r[id] = rate.NewLimiter(rate.Limit(limit), 0)
	})
}

// We limit up to 100 hits per site per day to ensure smooth operation but with
// enough traffic to stress test deployments.
//
// We don't care if it is bots or legit traffic, so long it is coming from a
// registered site we count it as a hit.
func limit() float64 {
	return float64(100) / (24 * time.Hour).Seconds()
}

func (c *Cache) Allow(domain string) (ok bool) {
	h := crc32.NewIEEE()
	h.Write([]byte(domain))
	id := h.Sum32()
	c.mu.RLock()
	ok = c.b.Contains(id) && c.r[id].Allow()
	c.mu.RUnlock()
	return
}

func (c *Cache) Add(domain string) {
	h := crc32.NewIEEE()
	h.Write([]byte(domain))
	id := h.Sum32()
	c.mu.Lock()
	c.b.Add(id)
	c.r[id] = rate.NewLimiter(rate.Limit(limit()), 0)
	c.mu.Unlock()
}

func (c *Cache) Remove(domain string) {
	h := crc32.NewIEEE()
	h.Write([]byte(domain))
	id := h.Sum32()
	c.mu.Lock()
	c.b.Remove(id)
	delete(c.r, id)
	c.mu.Unlock()
}
