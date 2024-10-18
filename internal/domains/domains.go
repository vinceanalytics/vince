package domains

import (
	"sync"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
)

var (
	domains = map[string]struct{}{}
	mu      sync.RWMutex
)

func Reload(l func(f func(*v1.Site))) {
	keys := make(map[string]struct{})
	l(func(s *v1.Site) {
		keys[s.Domain] = struct{}{}
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
