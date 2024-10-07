package domains

import (
	"sync/atomic"

	"github.com/RoaringBitmap/roaring/v2"
	"github.com/dgraph-io/badger/v4/y"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
)

var domains = newDom()

func newDom() *atomic.Pointer[roaring.Bitmap] {
	var d atomic.Pointer[roaring.Bitmap]
	d.Store(roaring.New())
	return &d
}

func Reload(l func(f func(*v1.Site))) {
	keys := roaring.New()
	l(func(s *v1.Site) {
		keys.Add(y.Hash([]byte(s.Domain)))
	})
	domains.Store(keys)
}

func Allow(domain string) bool {
	return domains.Load().
		Contains(y.Hash([]byte(domain)))
}
