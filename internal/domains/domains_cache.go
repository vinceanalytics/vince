package domains

import (
	"sync"

	"github.com/vinceanalytics/vince/internal/features"
)

type Cache struct {
	domains map[string]struct{}
	mu      sync.RWMutex
}

func New() *Cache {
	return &Cache{domains: make(map[string]struct{})}
}

func (c *Cache) Load() func(b []byte) {
	return func(b []byte) {
		c.domains[string(b)] = struct{}{}
	}
}

func (c *Cache) Allow(domain string) bool {
	c.mu.RLock()
	_, ok := c.domains[domain]
	c.mu.RUnlock()
	return ok && features.Allow()
}

func (c *Cache) Add(domain string) {
	c.mu.Lock()
	c.domains[domain] = struct{}{}
	c.mu.Unlock()
}

func (c *Cache) Remove(domain string) {
	c.mu.Lock()
	delete(c.domains, domain)
	c.mu.Unlock()
}
