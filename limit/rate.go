package limit

import (
	"sync"

	"golang.org/x/time/rate"
)

type Limit struct {
	m *sync.Map
}

func (l *Limit) Allow(id uint64, by rate.Limit, bust int) bool {
	x, ok := l.m.Load(id)
	if !ok {
		r := rate.NewLimiter(by, bust)
		l.m.Store(id, r)
		return r.Allow()
	}
	return x.(*rate.Limiter).Allow()
}

var API = &Limit{m: &sync.Map{}}
var SITES = &Limit{m: &sync.Map{}}
