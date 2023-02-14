package api

import (
	"crypto/rand"
	"sync"

	"github.com/dchest/siphash"
	"golang.org/x/net/publicsuffix"
)

type Hash struct {
	rand [16]byte
}

func (h *Hash) Random() *Hash {
	_, err := rand.Read(h.rand[:])
	if err != nil {
		panic("Failed to generate random seed " + err.Error())
	}
	return h
}

type ID struct {
	prev [16]byte
	h    [16]byte
	mu   sync.RWMutex
}

var seedID = (&ID{}).Setup()

func (i *ID) Reset() *ID {
	i.mu.Lock()
	defer i.mu.Unlock()
	copy(i.prev[:], i.h[:])
	_, err := rand.Read(i.h[:])
	if err != nil {
		panic("Failed to read random value" + err.Error())
	}
	return i
}

func (i *ID) Setup() *ID {
	rand.Read(i.prev[:])
	rand.Read(i.h[:])
	return i
}

func (i *ID) Gen(remoteIP, userAgent, domain, host string) uint64 {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.gen(i.h[:], remoteIP, userAgent, domain, host)
}

func (i *ID) GenPrevious(remoteIP, userAgent, domain, host string) uint64 {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.gen(i.prev[:], remoteIP, userAgent, domain, host)
}

func (i *ID) gen(h []byte, remoteIP, userAgent, domain, host string) uint64 {

	sh := siphash.New(h)
	rootDomain := host
	if s, ok := publicsuffix.PublicSuffix(host); ok {
		rootDomain = s
	}
	_, err := sh.Write([]byte(userAgent + remoteIP + domain + rootDomain))
	if err != nil {
		panic("Failed to hash user iD " + err.Error())
	}
	return sh.Sum64()
}
