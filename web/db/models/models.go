package models

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gernest/len64/web/db/schema"
	"gorm.io/gorm"
)

func UpdateSiteStartDate(db *gorm.DB, sid uint64, start time.Time) error {
	err := db.Model(&schema.Site{}).Where("id = ?", sid).Update("stats_start_date", start).Error
	if err != nil {
		return fmt.Errorf("update stats_start_date%w", err)
	}
	return nil
}

var domainRe = regexp.MustCompile(`(?P<domain>(?:[a-z0-9]+(?:-[a-z0-9]+)*\.)+[a-z]{2,})`)

func ValidateSiteDomain(db *gorm.DB, domain string) (good, bad string) {
	good = CleanupDOmain(domain)
	if good == "" {
		bad = "is required"
		return
	}
	if !domainRe.MatchString(good) {
		bad = "only letters, numbers, slashes and period allowed"
		return
	}
	if strings.ContainsAny(domain, reservedChars) {
		bad = "must not contain URI reserved characters " + reservedChars
		return
	}
	if schema.Exists(db, func(db *gorm.DB) *gorm.DB {
		return db.Model(&schema.Site{}).Where("domain = ?", domain)
	}) {
		bad = " already exists"
	}
	return
}

const reservedChars = `:?#[]@!$&'()*+,;=`

func CleanupDOmain(domain string) string {
	domain = strings.TrimSpace(domain)
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "www.")
	domain = strings.TrimSuffix(domain, "/")
	domain = strings.ToLower(domain)
	return domain
}
