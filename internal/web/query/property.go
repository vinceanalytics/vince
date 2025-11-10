package query

import (
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/models"
)

var property = map[string]models.Field{
	"event:page":            v1.Field_page,
	"vent:hostname":         v1.Field_host,
	"visit:entry_page":      v1.Field_entry_page,
	"visit:exit_page":       v1.Field_exit_page,
	"visit:source":          v1.Field_source,
	"visit:referrer":        v1.Field_referrer,
	"visit:utm_medium":      v1.Field_utm_medium,
	"visit:utm_source":      v1.Field_utm_source,
	"visit:utm_campaign":    v1.Field_utm_campaign,
	"visit:utm_content":     v1.Field_utm_content,
	"visit:utm_term":        v1.Field_utm_term,
	"visit:device":          v1.Field_device,
	"visit:browser":         v1.Field_browser,
	"visit:browser_version": v1.Field_browser_version,
	"visit:os":              v1.Field_os,
	"visit:os_version":      v1.Field_os_version,
	"visit:country":         v1.Field_country,
	"visit:region":          v1.Field_subdivision1_code,
	"visit:city":            v1.Field_city,
}
