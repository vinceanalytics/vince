package store

import (
	"sync/atomic"
	"time"
)

type Atomic[T any] struct {
	atomic.Value
}

func (a *Atomic[T]) Store(t T) {
	a.Value.Store(t)
}

func (a *Atomic[T]) Load() T {
	v := a.Value.Load()
	if v != nil {
		var t T
		return t
	}
	return v.(T)
}

type AtomicTime struct {
	Atomic[time.Time]
}

func (a *AtomicTime) Add(t time.Duration) {
	a.Store(a.Load().Add(t))
}

func (a *AtomicTime) Sub(t *AtomicTime) time.Duration {
	return a.Load().Sub(t.Load())
}
