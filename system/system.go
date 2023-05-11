package system

import (
	"github.com/prometheus/client_golang/prometheus"
)

// histograms
var (
	// buckets for seconds resolutions of histograms
	buckets      = []float64{.25, .5, 1, 2.5, 5, 10}
	SaveDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "vince",
			Name:      "events_save_duration_seconds",
			Help:      "Time taken to persist events to the storage.",
			Buckets:   buckets,
		},
	)
	DropSiteDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "vince",
			Name:      "site_drop_duration_seconds",
			Help:      "Time taken to permanently delete a site.",
			Buckets:   buckets,
		},
	)
	SiteCacheDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "vince",
			Name:      "site_cache_duration_seconds",
			Help:      "Time taken to load sites to cache.",
			Buckets:   buckets,
		},
	)
	CalendarReadDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "vince",
			Name:      "calendar_read_duration_seconds",
			Help:      "Time taken to read a calendar from storage.",
			Buckets:   buckets,
		},
	)
)

var (
	DataPoint = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "vince",
			Name:      "data_point",
			Help:      "Number of calls to /api/event",
		},
		[]string{"status"},
	)
	DataPointReceived = DataPoint.WithLabelValues("received")
	DataPointRejected = DataPoint.WithLabelValues("rejected")
	DataPointDropped  = DataPoint.WithLabelValues("dropped")
	DataPointAccepted = DataPoint.WithLabelValues("accepted")
)

// gauges
var (
	SitesInCache = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "vince",
		Name:      "sites_in_cache",
		Help:      "Active sites in memory cache.",
	})
)

func init() {
	prometheus.DefaultRegisterer.MustRegister(
		DataPoint,
		SitesInCache,
		SaveDuration,
		DropSiteDuration,
		SiteCacheDuration,
		CalendarReadDuration,
	)
}
