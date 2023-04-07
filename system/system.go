package system

import "sync"

var (
	// Distribution of save duration for each accumulated buffer
	SaveDuration      = &histogramMetric{name: "mike_save_duration"}
	MergeDuration     = &histogramMetric{name: "bob_merge_duration"}
	SiteCacheDuration = &histogramMetric{name: "sites_cache_update"}
)

var set = &sync.Map{}

func Get(name string) *histogramMetric {
	v, _ := set.LoadOrStore(name, new(histogramMetric))
	return v.(*histogramMetric)
}
