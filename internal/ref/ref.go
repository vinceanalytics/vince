package ref

import (
	"embed"
	"fmt"
	"io/fs"
	"math/rand/v2"
	"net/url"
	"strings"
	"sync"

	"github.com/dgraph-io/badger/v4/y"
	"github.com/vinceanalytics/vince/fb/ref"
	"github.com/vinceanalytics/vince/internal/roaring"
)

//go:embed favicon
var faviconData embed.FS

var Favicon, _ = fs.Sub(faviconData, "favicon")

//go:generate go run gen/main.go

//go:embed refs.fbs.bin
var data []byte

type Re struct {
	root  *ref.Ref
	bsi   *roaring.BSI
	mu    sync.RWMutex
	cache map[uint32][]byte
}

var base = New()

func Search(uri string) ([]byte, error) {
	return base.Search(uri)
}

func New() *Re {
	root := ref.GetRootAsRef(data, 0)
	bsi := roaring.NewBSIFromBuffer(root.BsiBytes())

	return &Re{
		root:  root,
		bsi:   bsi,
		cache: make(map[uint32][]byte),
	}
}

func (r *Re) Search(uri string) ([]byte, error) {
	base, err := clean(uri)
	if err != nil {
		return nil, err
	}
	key := []byte(base)
	hash := y.Hash(key)
	r.mu.RLock()
	cached, ok := r.cache[hash]
	r.mu.RUnlock()
	if ok {
		return cached, nil
	}
	idx, ok := r.bsi.GetValue(uint64(hash))
	if !ok {
		return key, nil
	}
	value := r.root.Ref(int(idx))
	r.mu.Lock()
	r.cache[hash] = value
	r.mu.Unlock()
	return value, nil
}

func clean(r string) (string, error) {
	if strings.HasPrefix(r, "http://") || strings.HasPrefix(r, "https://") {
		u, err := url.Parse(r)
		if err != nil {
			return "", fmt.Errorf("cleaning referer uri%w", err)
		}
		return u.Host, nil
	}
	return r, nil
}

// Rand returns a random referrer. used in generating seeds dvents.
func Rand() string {
	n := rand.IntN(base.root.RefLength())
	return string(base.root.Ref(n))
}
