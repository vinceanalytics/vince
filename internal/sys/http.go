package sys

import (
	"net/http"
	"sync/atomic"
	"time"
)

var (
	// <= 0.5 seconds
	b0 atomic.Uint64
	// <= 1 second
	b1 atomic.Uint64
	// > 1 second
	b2 atomic.Uint64
)

const (
	ms0 = 500 * time.Millisecond
	ms1 = 1000 * time.Millisecond
)

func HTTP(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, r)
		elapsed := time.Since(start)
		if elapsed <= ms0 {
			b0.Add(1)
		} else if elapsed <= ms1 {
			b1.Add(1)
		} else {
			b2.Add(1)
		}
	})
}
