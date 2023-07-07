package models

import (
	"context"

	"github.com/vinceanalytics/vince/pkg/schema"
	"golang.org/x/time/rate"
)

type CachedSite = schema.CachedSite

func CacheRateLimit(c *CachedSite) (rate.Limit, int) {
	if !c.IngestRateLimit.Valid {
		return rate.Inf, 0
	}
	return rate.Limit(c.IngestRateLimit.Float64), 10
}

func QuerySitesToCache(ctx context.Context, fn func(*CachedSite)) (count float64) {
	db := Get(ctx)
	var sites []*CachedSite
	err := db.Model(&Site{}).Select(
		"id", "domain", "stats_start_date", "ingest_rate_limit", "user_id",
	).Find(&sites).Error
	if err != nil {
		LOG(ctx, err, "failed getting sites to cache")
	}
	for _, s := range sites {
		fn(s)
	}
	return
}
