package domains

import (
	"sync"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
)

var domains = map[string]struct{}{}
var mu sync.RWMutex

type Loader func(func(*v1.Site))

func Load(l Loader) {
	mu.Lock()
	l(load)
	mu.Unlock()
}

func Reload(l Loader) {
	mu.Lock()
	clear(domains)
	l(load)
	mu.Unlock()
}

func load(s *v1.Site) {
	if !s.Locked {
		domains[s.Domain] = struct{}{}
	}
}

func Allow(domain string) bool {
	mu.RLock()
	_, ok := domains[domain]
	mu.RUnlock()
	return ok
}

func Add(domain string) {
	mu.Lock()
	domains[domain] = struct{}{}
	mu.Unlock()
}

func Remove(domain string) {
	mu.Lock()
	delete(domains, domain)
	mu.Unlock()
}
