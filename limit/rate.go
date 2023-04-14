package limit

import (
	"sync"

	"golang.org/x/time/rate"
)

type Limit struct {
	m *sync.Map
}

func (l *Limit) AllowID(id uint64, by rate.Limit, burst int) bool {
	x, ok := l.m.Load(id)
	if !ok {
		x := rate.NewLimiter(by, burst)
		l.m.Store(id, x)
		return x.Allow()
	}
	return x.(*rate.Limiter).Allow()
}

var API = &Limit{m: &sync.Map{}}
