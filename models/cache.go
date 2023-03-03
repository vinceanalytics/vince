package models

import (
	"errors"
	"time"

	"golang.org/x/time/rate"
	"gorm.io/gorm"
)

type CachedSite struct {
	ID                          uint64
	Domain                      string
	IngestRateLimitScaleSeconds uint64
	IngestRateLimitThreshold    uint64
	UserID                      uint64
}

func (c *CachedSite) RateLimit() (uint64, rate.Limit, int) {
	events := float64(c.IngestRateLimitThreshold)
	per := time.Duration(c.IngestRateLimitScaleSeconds) * time.Second
	return c.ID, rate.Limit(events / per.Seconds()), 10
}

func QuerySitesToCache(db *gorm.DB, results *[]*CachedSite) error {
	err := db.Model(&Site{}).Select("sites.id, sites.domain, sites.ingest_rate_limit_scale_seconds,sites.ingest_rate_limit_threshold,site_memberships.user_id").
		Joins("left join  site_memberships on sites.id = site_memberships.site_id and site_memberships.role = 'owner' ").
		Scan(results).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	return nil
}
