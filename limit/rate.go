package limit

import (
	"sync"

	"github.com/gernest/vince/models"
	"golang.org/x/time/rate"
)

type siteRate struct {
	cached *models.CachedSite
	rate   *rate.Limiter
}

type Limit struct {
	m *sync.Map
}

func (l *Limit) Allow(domain string) (*models.CachedSite, bool) {
	x, ok := l.m.Load(domain)
	if !ok {
		return nil, false
	}
	v := x.(*siteRate)
	return v.cached, v.rate.Allow()
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

func (l *Limit) Set(m *models.CachedSite) {
	l.m.Store(m.Domain, &siteRate{
		cached: m,
		rate:   rate.NewLimiter(m.RateLimit()),
	})
}

var API = &Limit{m: &sync.Map{}}
var SITES = &Limit{m: &sync.Map{}}
