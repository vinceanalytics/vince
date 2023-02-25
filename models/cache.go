package models

import "gorm.io/gorm"

type CachedSite struct {
	ID                          uint64
	Domain                      string
	IngestRateLimitScaleSeconds uint64
	IngestRateLimitThreshold    uint64
	UserID                      uint64
}

func QuerySitesToCache(db *gorm.DB) ([]*CachedSite, error) {
	var results []*CachedSite
	err := db.Model(&Site{}).Select("sites.id, sites.domain, sites.ingest_rate_limit_scale_seconds,sites.ingest_rate_limit_threshold,site_memberships.user_id").
		Joins("left join  site_memberships on sites.id = site_memberships.site_id and site_memberships.role = 'owner' ").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}
