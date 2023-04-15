package system

import (
	"runtime"
	"time"
)

// histograms
var (
	SaveDuration      = &histogramMetric{name: "mike_save_duration"}
	MergeDuration     = &histogramMetric{name: "bob_merge_duration"}
	SiteCacheDuration = &histogramMetric{name: "sites_cache_update"}
)

// counters for the /api/event endpoint.
var (
	// DataPointReceived is cumulative total of calls to /api/event
	DataPointReceived = &counterMetric{name: "data_point_received_total"}
	// DataPointReceived is cumulative total of calls to /api/event that were successfully
	// accepted and processed.
	DataPointAccepted = &counterMetric{name: "data_point_accepted_total"}
	// DataPointReceived is cumulative total of calls to /api/event that were rejected with
	// bad request. We don't care about the actual reason, we need only an indication of something
	// was not right.
	DataPointRejectedBadRequest = &counterMetric{name: "data_point_rejected_bad_request_total"}
	// DataPointReceived is cumulative total of calls to /api/event that were dropped. This happens
	// when calls are made to deleted sites or sites with exceeded limits/quotas.
	DataPointDropped = &counterMetric{name: "data_point_dropped_total"}
)

// gauges
var (
	SitesInCache = &gaugeMetric{name: "sites_in_cache"}
)

var syncStats = &Sync{}

type Sync struct {
	ms    runtime.MemStats
	stats Stats
}

func (s *Sync) read(ts time.Time) {
	s.readGo(ts)
}

func (s *Sync) Collect(ts time.Time) Stats {
	s.read(ts)
	return s.stats
}

func Collect(ts time.Time) Stats {
	return syncStats.Collect(ts)
}

func CollectHist(ts time.Time) []*Histogram {
	return []*Histogram{
		SaveDuration.Read(ts),
		MergeDuration.Read(ts),
		SiteCacheDuration.Read(ts),
	}
}
