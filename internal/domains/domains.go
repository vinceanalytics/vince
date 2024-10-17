package domains

import (
	"sync"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
)

var (
	domains = map[string]uint64{}
	mu      sync.RWMutex
)

func Reload(l func(f func(*v1.Site))) {
	keys := make(map[string]uint64)
	l(func(s *v1.Site) {
		keys[s.Domain] = s.Id
	})
	mu.Lock()
	domains = keys
	mu.Unlock()
}

func Allow(domain string) (ok bool) {
	mu.RLock()
	_, ok = domains[domain]
	mu.RUnlock()
	return
}

func Count() (n uint64) {
	mu.RLock()
	n = uint64(len(domains))
	mu.RUnlock()
	return
}

func ID(domain string) (n uint64) {
	mu.RLock()
	n = domains[domain]
	mu.RUnlock()
	return
}
