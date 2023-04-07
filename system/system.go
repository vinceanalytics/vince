package system

import "sync"

var (
	// Distribution of save duration for each accumulated buffer
	SaveDuration      = Get("mike_save_duration")
	MergeDuration     = Get("bob_merge_duration")
	SiteCacheDuration = Get("sites_cache_update")
)

var set = &sync.Map{}

func Get(name string) *Histogram {
	v, _ := set.LoadOrStore(name, new(Histogram))
	return v.(*Histogram)
}

type Reader interface {
	Read(*History)
}

func All(f func(Reader)) {
	set.Range(func(key, value any) bool {
		f(value.(*Histogram))
		return true
	})
}
