package query

import "github.com/vinceanalytics/vince/internal/models"

var property = map[string]models.Field{
	"event:page":            models.Field_page,
	"vent:hostname":         models.Field_host,
	"visit:entry_page":      models.Field_entry_page,
	"visit:exit_page":       models.Field_exit_page,
	"visit:source":          models.Field_source,
	"visit:referrer":        models.Field_referrer,
	"visit:utm_medium":      models.Field_utm_medium,
	"visit:utm_source":      models.Field_utm_source,
	"visit:utm_campaign":    models.Field_utm_campaign,
	"visit:utm_content":     models.Field_utm_content,
	"visit:utm_term":        models.Field_utm_term,
	"visit:device":          models.Field_device,
	"visit:browser":         models.Field_browser,
	"visit:browser_version": models.Field_browser_version,
	"visit:os":              models.Field_os,
	"visit:os_version":      models.Field_os_version,
	"visit:country":         models.Field_country,
	"visit:region":          models.Field_subdivision1_code,
	"visit:city":            models.Field_city,
}
