package limit

import (
	"sync"

	"golang.org/x/time/rate"
)

type Limit struct {
	m *sync.Map
}

func (l *Limit) Allow(id uint64, by rate.Limit, bust int) bool {
	x, _ := l.m.LoadOrStore(id, rate.NewLimiter(by, bust))
	return x.(*rate.Limiter).Allow()
}

var API = &Limit{m: &sync.Map{}}
var SITES = &Limit{m: &sync.Map{}}
