package timeseries

import (
	"strings"
	"time"

	"github.com/segmentio/parquet-go"
)

type Metrics struct {
	Timestamp      time.Time        `parquet:"timestamp"`
	Visitors       uint64           `parquet:"visitors,zstd"`
	Visits         uint64           `parquet:"visits,zstd"`
	PageViews      uint64           `parquet:"page_views,zstd"`
	ViewsPerVisit  float64          `parquet:"views_per_visit,zstd"`
	VisitDuration  time.Duration    `parquet:"visit_duration,zstd"`
	Events         uint64           `parquet:"events,zstd"`
	Name           map[string]Total `parquet:"name" parquet-key:",dict,zstd"`
	Page           map[string]Total `parquet:"page" parquet-key:",dict,zstd"`
	EntryPage      map[string]Total `parquet:"entry_page" parquet-key:",dict,zstd"`
	ExitPage       map[string]Total `parquet:"exit_page" parquet-key:",dict,zstd"`
	Referrer       map[string]Total `parquet:"referrer" parquet-key:",dict,zstd"`
	UtmMedium      map[string]Total `parquet:"utm_medium" parquet-key:",dict,zstd"`
	UtmSource      map[string]Total `parquet:"utm_source" parquet-key:",dict,zstd"`
	UtmCampaign    map[string]Total `parquet:"utm_campaign" parquet-key:",dict,zstd"`
	UtmContent     map[string]Total `parquet:"utm_content" parquet-key:",dict,zstd"`
	UtmTerm        map[string]Total `parquet:"utm_term" parquet-key:",dict,zstd"`
	UtmDevice      map[string]Total `parquet:"utm_device" parquet-key:",dict,zstd"`
	Browser        map[string]Total `parquet:"browser" parquet-key:",dict,zstd"`
	BrowserVersion map[string]Total `parquet:"browser_version" parquet-key:",dict,zstd"`
	Os             map[string]Total `parquet:"os" parquet-key:",dict,zstd"`
	OsVersion      map[string]Total `parquet:"os_version" parquet-key:",dict,zstd"`
	Country        map[string]Total `parquet:"country" parquet-key:",dict,zstd"`
	Region         map[string]Total `parquet:"region" parquet-key:",dict,zstd"`
	City           map[string]Total `parquet:"city" parquet-key:",dict,zstd"`
}

type Total struct {
	Visitors      uint64        `parquet:"visitors,zstd"`
	Visits        uint64        `parquet:"visits,zstd"`
	PageViews     uint64        `parquet:"page_views,zstd"`
	ViewsPerVisit float64       `parquet:"views_per_visit,zstd"`
	VisitDuration time.Duration `parquet:"visit_duration,zstd"`
	Events        uint64        `parquet:"events,zstd"`
}

var schema = parquet.SchemaOf(&Metrics{})

// Given pro and metric returns LeafColumn for prop and metric  inside the  parquet
// schema.
//
// Use this to read property columns.
func lookup(prop PROPS, metric ...METRIC_TYPE) (key parquet.LeafColumn, values []parquet.LeafColumn) {
	k := strings.ToLower(prop.String())
	key, _ = schema.Lookup(k, "key_value", "key")
	values = make([]parquet.LeafColumn, len(metric))
	for i := 0; i < len(metric); i++ {
		x, _ := schema.Lookup(k, "key_value", "value", metric[i].String())
		values[i] = x
	}
	return
}
