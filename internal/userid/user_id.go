package userid

import (
	"sync"

	"github.com/cespare/xxhash/v2"
)

func Hash(remoteIP, userAgent, domain, host string) (r uint64) {
	h := get()
	h.WriteString(remoteIP)
	h.WriteString(userAgent)
	h.WriteString(domain)
	h.WriteString(host)
	r = h.Sum64()
	put(h)
	return
}

var digestPool = &sync.Pool{
	New: func() any {
		return xxhash.New()
	},
}

func get() *xxhash.Digest {
	return digestPool.Get().(*xxhash.Digest)
}

func put(h *xxhash.Digest) {
	h.Reset()
	digestPool.Put(h)
}
