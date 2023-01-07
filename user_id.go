package vince

import (
	"crypto/rand"
	"sync"

	"github.com/dchest/siphash"
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
	h  [16]byte
	mu sync.RWMutex
}

var seedID = (&ID{}).Reset()

func (i *ID) Reset() *ID {
	i.mu.Lock()
	defer i.mu.Unlock()
	_, err := rand.Read(i.h[:])
	if err != nil {
		panic("Failed to read random value" + err.Error())
	}
	return i
}

func (i *ID) Gen(remoteIP, userAgent, domain string) uint64 {
	i.mu.RLock()
	defer i.mu.RUnlock()
	sh := siphash.New(i.h[:])
	_, err := sh.Write([]byte(remoteIP + userAgent + domain))
	if err != nil {
		panic("Failed to hash user iD " + err.Error())
	}
	return sh.Sum64()
}
