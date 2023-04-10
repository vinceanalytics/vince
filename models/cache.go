package models

import (
	"context"
	"time"

	"golang.org/x/time/rate"
)

type CachedSite struct {
	ID                          uint64
	Domain                      string
	IngestRateLimitScaleSeconds uint64
	IngestRateLimitThreshold    uint64
	UserID                      uint64
}

func (c *CachedSite) RateLimit() (rate.Limit, int) {
	events := float64(c.IngestRateLimitThreshold)
	per := time.Duration(c.IngestRateLimitScaleSeconds) * time.Second
	return rate.Limit(events / per.Seconds()), 10
}

func QuerySitesToCache(ctx context.Context, fn func(*CachedSite)) (count float64) {
	db := Get(ctx)
	rows, err := db.Model(&Site{}).Select("sites.id, sites.domain, sites.ingest_rate_limit_scale_seconds,sites.ingest_rate_limit_threshold,site_memberships.user_id").
		Joins("left join  site_memberships on sites.id = site_memberships.site_id and site_memberships.role = 'owner' ").
		Rows()
	if err != nil {
		DBE(ctx, err, "failed getting sites to cache")
	} else {
		var site CachedSite
		for rows.Next() {
			db.ScanRows(rows, &site)
			fn(&site)
			count += 1
		}
	}
	return
}
