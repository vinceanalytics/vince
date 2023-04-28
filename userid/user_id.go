package userid

import (
	"context"
	"crypto/rand"
	"hash"
	"sync"

	"github.com/dchest/siphash"
	"golang.org/x/net/publicsuffix"
)

type ID struct {
	prev                  [16]byte
	h                     [16]byte
	hashCurrent, hashPrev hash.Hash64
	mu                    sync.RWMutex
}

type userIDKey struct{}

func Open(ctx context.Context) context.Context {
	return context.WithValue(ctx, userIDKey{}, (&ID{}).Setup())
}

func Get(ctx context.Context) *ID {
	return ctx.Value(userIDKey{}).(*ID)
}

func (i *ID) Reset() {
	i.mu.Lock()
	defer i.mu.Unlock()
	copy(i.prev[:], i.h[:])
	_, err := rand.Read(i.h[:])
	if err != nil {
		panic("Failed to read random value" + err.Error())
	}
	i.hashCurrent = siphash.New(i.h[:])
	i.hashPrev = siphash.New(i.prev[:])
	return
}

func (i *ID) Setup() *ID {
	rand.Read(i.prev[:])
	rand.Read(i.h[:])
	i.hashCurrent = siphash.New(i.h[:])
	i.hashPrev = siphash.New(i.prev[:])
	return i
}

func (i *ID) Hash(remoteIP, userAgent, domain, host string) uint64 {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.gen(i.hashCurrent, remoteIP, userAgent, domain, host)
}

func (i *ID) HashPrevious(remoteIP, userAgent, domain, host string) uint64 {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.gen(i.hashPrev, remoteIP, userAgent, domain, host)
}

func (i *ID) gen(h hash.Hash64, remoteIP, userAgent, domain, host string) uint64 {
	h.Reset()
	rootDomain := host
	if s, ok := publicsuffix.PublicSuffix(host); ok {
		rootDomain = s
	}
	_, err := h.Write([]byte(userAgent + remoteIP + domain + rootDomain))
	if err != nil {
		panic("Failed to hash user iD " + err.Error())
	}
	return h.Sum64()
}
