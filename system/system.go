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

var syncStats = &Sync{
	metrics: metrics{
		histograms: []*histogramMetric{
			SaveDuration, MergeDuration, SiteCacheDuration,
		},
		gauges: []*gaugeMetric{
			SitesInCache,
		},
		counters: []*counterMetric{
			DataPointReceived, DataPointAccepted, DataPointRejectedBadRequest, DataPointDropped,
		},
	},
	gauges:     make([]*Gauge, 0, 1024),
	counters:   make([]*Counter, 0, 1024),
	histograms: make([]*Histogram, 0, 1024),
}

type metrics struct {
	gauges     []*gaugeMetric
	counters   []*counterMetric
	histograms []*histogramMetric
}

type Sync struct {
	metrics    metrics
	ms         runtime.MemStats
	gauges     []*Gauge
	counters   []*Counter
	histograms []*Histogram
}

func (s *Sync) read(ts time.Time) {
	s.counters = s.counters[:0]
	s.gauges = s.gauges[:0]
	s.histograms = s.histograms[:0]
	for _, m := range s.metrics.gauges {
		s.gauges = append(s.gauges, m.Read(ts))
	}
	for _, m := range s.metrics.counters {
		s.counters = append(s.counters, m.Read(ts))
	}
	for _, m := range s.metrics.histograms {
		s.histograms = append(s.histograms, m.Read(ts))
	}
	s.readGo(ts)
}

func (s *Sync) Collect(ts time.Time, co Collector) {
	s.read(ts)
	co.Gauges(s.gauges)
	co.Counters(s.counters)
	co.Histograms(s.histograms)
}

type Collector struct {
	Gauges     func([]*Gauge)
	Counters   func([]*Counter)
	Histograms func([]*Histogram)
}

func Collect(ts time.Time, co Collector) {
	syncStats.Collect(ts, co)
}
