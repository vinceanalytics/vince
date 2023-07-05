package system

import (
	"context"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/vinceanalytics/vince/internal/core"
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
	QueryDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "vince",
			Name:      "events_query_duration_seconds",
			Help:      "Time taken to query.",
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

	KeysPerBufferSave = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "vince",
			Name:      "ts_key_per_buffer_save",
			Help:      "Time number of keys created on a single timeseries buffer saving.",
			Buckets:   []float64{50, 100, 150, 200, 500},
		},
	)
	EntriesPerBufferSave = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "vince",
			Name:      "ts_entries_per_buffer_save",
			Help:      "Time number of entries buffered on a single buffer saving.",
			Buckets:   []float64{10, 20, 50, 100},
		},
	)
)

var (
	DataPointReceived = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "vince",
		Name:      "data_point_received",
	})
	DataPointRejected = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "vince",
		Name:      "data_point_rejected",
	})
	DataPointDropped = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "vince",
		Name:      "data_point_dropped",
	})
	DataPointAccepted = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "vince",
		Name:      "data_point_accepted",
	})
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
		DataPointReceived,
		DataPointRejected,
		DataPointDropped,
		DataPointAccepted,
		SitesInCache,
		SaveDuration,
		DropSiteDuration,
		SiteCacheDuration,
		CalendarReadDuration,
	)
}

type Stats struct {
	Timestamp         time.Time `parquet:"timestamp,timestamp"`
	SitesInCache      int64     `parquet:"sites_in_cache"`
	DataPointReceived int64     `parquet:"data_point_received,zstd"`
	DataPointRejected int64     `parquet:"data_point_rejected,zstd"`
	DataPointDropped  int64     `parquet:"data_point_dropped,zstd"`
	DataPointAccepted int64     `parquet:"data_point_accepted,zstd"`
	TotalAllocation   int64     `parquet:"total_allocation,zstd"`
}

func (s *Stats) Read(ts time.Time) {
	m := new(dto.Metric)

	SitesInCache.Write(m)
	s.SitesInCache = int64(m.GetGauge().GetValue())

	m.Reset()
	DataPointReceived.Write(m)
	s.DataPointReceived = int64(m.GetCounter().GetValue())

	m.Reset()
	DataPointRejected.Write(m)
	s.DataPointRejected = int64(m.GetCounter().GetValue())

	m.Reset()
	DataPointDropped.Write(m)
	s.DataPointDropped = int64(m.GetCounter().GetValue())

	m.Reset()
	DataPointAccepted.Write(m)
	s.DataPointAccepted = int64(m.GetCounter().GetValue())

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	s.TotalAllocation = int64(mem.TotalAlloc)
}

func Read(ctx context.Context) (o Stats) {
	o.Read(core.Now(ctx))
	return
}
