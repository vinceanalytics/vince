package domains

import (
	"sync"

	"github.com/dgraph-io/badger/v4/y"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/roaring"
)

var domains = roaring.New()

var mu sync.RWMutex

type Loader func(f func(*v1.Site))

func Reload(l Loader) {
	keys := roaring.New()
	l(func(s *v1.Site) {
		keys.Add(y.Hash([]byte(s.Domain)))
	})
	mu.Lock()
	domains = keys
	mu.Unlock()
}

func Allow(domain string) bool {
	mu.RLock()
	ok := domains.Contains(y.Hash([]byte(domain)))
	mu.RUnlock()
	return ok
}
