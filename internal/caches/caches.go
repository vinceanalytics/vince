package caches

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/pkg/entry"
	"golang.org/x/time/rate"
)

type sessionKey struct{}
type sitesKey struct{}
type userKey struct{}
type ipKey struct{}
type apiKey struct{}

func Open(ctx context.Context, onSession func(context.Context, *entry.Entry)) (context.Context, error) {
	session, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     2 << 20,
		BufferItems: 64,
		OnEvict: func(item *ristretto.Item) {
			onSession(ctx, item.Value.(*entry.Entry))
		},
	})
	if err != nil {
		return nil, err
	}
	sites, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 20,
		BufferItems: 64,
	})
	if err != nil {
		session.Close()
		return nil, err
	}
	users, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 20,
		BufferItems: 64,
	})
	if err != nil {
		session.Close()
		sites.Close()
		return nil, err
	}
	ip, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 20,
		BufferItems: 64,
	})
	if err != nil {
		session.Close()
		sites.Close()
		users.Close()
		return nil, err
	}
	api, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 20,
		BufferItems: 64,
	})
	if err != nil {
		session.Close()
		sites.Close()
		users.Close()
		ip.Close()
		return nil, err
	}

	if err != nil {
		session.Close()
		sites.Close()
		users.Close()
		ip.Close()
		api.Close()
		return nil, err
	}
	ctx = context.WithValue(ctx, sessionKey{}, session)
	ctx = context.WithValue(ctx, sitesKey{}, sites)
	ctx = context.WithValue(ctx, userKey{}, users)
	ctx = context.WithValue(ctx, ipKey{}, ip)
	ctx = context.WithValue(ctx, apiKey{}, api)
	return ctx, nil
}

func Close(ctx context.Context) error {
	Session(ctx).Close()
	Site(ctx).Close()
	User(ctx).Close()
	IP(ctx).Close()
	API(ctx).Close()
	return nil
}

func Session(ctx context.Context) *ristretto.Cache {
	return ctx.Value(sessionKey{}).(*ristretto.Cache)
}

func Site(ctx context.Context) *ristretto.Cache {
	return ctx.Value(sitesKey{}).(*ristretto.Cache)
}

func User(ctx context.Context) *ristretto.Cache {
	return ctx.Value(userKey{}).(*ristretto.Cache)
}

func IP(ctx context.Context) *ristretto.Cache {
	return ctx.Value(ipKey{}).(*ristretto.Cache)
}

func API(ctx context.Context) *ristretto.Cache {
	return ctx.Value(apiKey{}).(*ristretto.Cache)
}

type SiteRate struct {
	SID        uint64
	UID        uint64
	HasStarted atomic.Bool
	Rate       *rate.Limiter
}

func (s *SiteRate) Allow(ctx context.Context) (uint64, uint64, bool) {
	ok := s.Rate.Allow()
	if ok {
		// we have allowed this event tp be processed. We need to update site with
		// the date which we accepted the first event
		if !s.HasStarted.Load() {
			models.UpdateSiteStartDate(ctx, s.SID, core.Now(ctx))
			s.HasStarted.Store(true)
		}
	}
	return s.UID, s.SID, ok
}

func SetSite(ctx context.Context, ttl time.Duration) func(*models.CachedSite) {
	cache := Site(ctx)
	return func(cs *models.CachedSite) {
		s := &SiteRate{
			SID:  cs.ID,
			UID:  cs.UserID,
			Rate: rate.NewLimiter(models.CacheRateLimit(cs)),
		}
		s.HasStarted.Store(cs.StatsStartDate.IsZero())
		cache.SetWithTTL(cs.Domain, s, 1, ttl)
	}
}

func AllowSite(ctx context.Context, domain string) (uid, sid uint64, ok bool) {
	x, _ := Site(ctx).Get(domain)
	if x != nil {
		r := x.(*SiteRate)
		return r.Allow(ctx)
	}
	return
}

// LoginRate limit to 2 login attempts per second.
var LoginRate = rate.Limit(2.0)

func AllowUseIDToLogin(ctx context.Context, uid uint64) bool {
	r := User(ctx)
	x, ok := r.Get(uid)
	if !ok {
		r.Set(uid, rate.NewLimiter(LoginRate, 10), 1)
		return true
	}
	return x.(*rate.Limiter).Allow()
}

func AllowRemoteIPLogin(ctx context.Context, ip string) bool {
	r := IP(ctx)
	x, ok := r.Get(ip)
	if !ok {
		r.Set(ip, rate.NewLimiter(LoginRate, 10), 1)
		return true
	}
	return x.(*rate.Limiter).Allow()
}

func AllowAPI(ctx context.Context, aid uint64, by rate.Limit, burst int) bool {
	r := API(ctx)
	x, ok := r.Get(aid)
	if !ok {
		x := rate.NewLimiter(by, burst)
		r.Set(aid, x, 1)
		return x.Allow()
	}
	return x.(*rate.Limiter).Allow()
}
