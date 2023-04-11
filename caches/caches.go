package caches

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/timex"
	"golang.org/x/time/rate"
)

type sessionKey struct{}
type sitesKey struct{}

func Open(ctx context.Context) (context.Context, error) {
	session, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     2 << 20,
		BufferItems: 64,
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
	ctx = context.WithValue(ctx, sessionKey{}, session)
	ctx = context.WithValue(ctx, sitesKey{}, sites)
	return ctx, nil
}

func Close(ctx context.Context) {
	Session(ctx).Close()
	Site(ctx).Close()
}

func Session(ctx context.Context) *ristretto.Cache {
	return ctx.Value(sessionKey{}).(*ristretto.Cache)
}

func Site(ctx context.Context) *ristretto.Cache {
	return ctx.Value(sitesKey{}).(*ristretto.Cache)
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
			models.UpdateSiteStartDate(ctx, s.SID, timex.Today())
			s.HasStarted.Store(true)
		}
	}
	return s.UID, s.SID, ok
}

func SetSite(ctx context.Context, ttl time.Duration) func(*models.CachedSite) {
	cache := Site(ctx)
	return func(cs *models.CachedSite) {
		var ok atomic.Bool
		ok.Store(!cs.StatsStartDate.IsZero())
		cache.SetWithTTL(cs.Domain, &SiteRate{
			SID:        cs.ID,
			UID:        cs.UserID,
			HasStarted: ok,
			Rate:       rate.NewLimiter(cs.RateLimit()),
		}, 1, ttl)
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
