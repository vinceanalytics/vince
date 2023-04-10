package limit

import (
	"sync"

	"github.com/gernest/vince/models"
	"golang.org/x/time/rate"
)

type siteRate struct {
	sid  uint64
	uid  uint64
	rate *rate.Limiter
}

type Limit struct {
	m *sync.Map
}

func (l *Limit) Allow(domain string) (uid, sid uint64, ok bool) {
	x, ok := l.m.Load(domain)
	if !ok {
		return
	}
	v := x.(*siteRate)
	return v.uid, v.sid, v.rate.Allow()
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
		sid:  m.ID,
		uid:  m.UserID,
		rate: rate.NewLimiter(m.RateLimit()),
	})
}

var API = &Limit{m: &sync.Map{}}
