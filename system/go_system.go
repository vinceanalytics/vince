package system

import (
	"runtime"
	"time"
)

type Stats struct {
	Timestamp     time.Time `parquet:"timestamp"`
	Alloc         float64   `parquet:"go_memstats_alloc_bytes"`
	TotalAlloc    float64   `parquet:"go_memstats_alloc_bytes_total"`
	Sys           float64   `parquet:"go_memstats_sys_bytes"`
	Lookups       float64   `parquet:"go_memstats_lookups_total"`
	Mallocs       float64   `parquet:"go_memstats_mallocs_total"`
	Frees         float64   `parquet:"go_memstats_frees_total"`
	HeapAlloc     float64   `parquet:"go_memstats_heap_alloc_bytes"`
	HeapSys       float64   `parquet:"go_memstats_heap_sys_bytes"`
	HeapIdle      float64   `parquet:"go_memstats_heap_idle_bytes"`
	HeapInuse     float64   `parquet:"go_memstats_heap_inuse_bytes"`
	HeapReleased  float64   `parquet:"go_memstats_heap_released_bytes"`
	HeapObjects   float64   `parquet:"go_memstats_heap_objects"`
	StackInuse    float64   `parquet:"go_memstats_stack_inuse_bytes"`
	StackSys      float64   `parquet:"go_memstats_stack_sys_bytes"`
	MSpanInuse    float64   `parquet:"go_memstats_mspan_inuse_bytes"`
	MSpanSys      float64   `parquet:"go_memstats_mspan_sys_bytes"`
	MCacheInuse   float64   `parquet:"go_memstats_mcache_inuse_bytes"`
	MCacheSys     float64   `parquet:"go_memstats_mcache_sys_bytes"`
	BuckHashSys   float64   `parquet:"go_memstats_alloc_bytes_total"`
	GCSys         float64   `parquet:"go_memstats_gc_sys_bytes"`
	OtherSys      float64   `parquet:"go_memstats_other_sys_bytes"`
	NextGC        float64   `parquet:"go_memstats_next_gc_bytes"`
	LastGC        float64   `parquet:"go_memstats_last_gc_time_seconds"`
	PauseTotalNs  float64   `parquet:"go_gc_duration_seconds_sum"`
	NumGC         float64   `parquet:"go_gc_duration_seconds_count"`
	NumForcedGC   float64   `parquet:"go_gc_forced_count"`
	GCCPUFraction float64   `parquet:"go_memstats_gc_cpu_fraction"`
	CGOCalls      float64   `parquet:"go_cgo_calls_count"`
	CPU           float64   `parquet:"go_cpu_count"`
	Goroutines    float64   `parquet:"go_goroutines"`
	MaxProcs      float64   `parquet:"go_gomaxprocs"`
	Threads       float64   `parquet:"go_threads"`

	// counters
	DataPointReceived           float64 `parquet:"data_point_received_total"`
	DataPointAccepted           float64 `parquet:"data_point_accepted_total"`
	DataPointRejectedBadRequest float64 `parquet:"data_point_rejected_bad_request_total"`
	DataPointDropped            float64 `parquet:"data_point_dropped_total"`

	// gauges
	SitesInCache float64 `parquet:"sites_in_cache"`
}

func (m *Stats) From(o *runtime.MemStats) {
	numThread, _ := runtime.ThreadCreateProfile(nil)
	m.Alloc = float64(o.Alloc)
	m.TotalAlloc = float64(o.TotalAlloc)
	m.Sys = float64(o.Sys)
	m.Lookups = float64(o.Lookups)
	m.Mallocs = float64(o.Mallocs)
	m.Frees = float64(o.Frees)
	m.HeapAlloc = float64(o.HeapAlloc)
	m.HeapSys = float64(o.HeapSys)
	m.HeapIdle = float64(o.HeapIdle)
	m.HeapInuse = float64(o.HeapInuse)
	m.HeapReleased = float64(o.HeapReleased)
	m.HeapObjects = float64(o.HeapObjects)
	m.StackInuse = float64(o.StackInuse)
	m.StackSys = float64(o.StackSys)
	m.MSpanInuse = float64(o.MSpanInuse)
	m.MSpanSys = float64(o.MSpanSys)
	m.MCacheInuse = float64(o.MCacheInuse)
	m.MCacheSys = float64(o.MCacheSys)
	m.BuckHashSys = float64(o.BuckHashSys)
	m.GCSys = float64(o.GCSys)
	m.OtherSys = float64(o.OtherSys)
	m.NextGC = float64(o.NextGC)
	m.LastGC = float64(o.LastGC) / 1e9
	m.PauseTotalNs = float64(o.PauseTotalNs) / 1e9
	m.NumGC = float64(o.NumGC)
	m.NumForcedGC = float64(o.NumForcedGC)
	m.GCCPUFraction = o.GCCPUFraction
	m.MaxProcs = float64(runtime.GOMAXPROCS(0))
	m.Goroutines = float64(runtime.NumGoroutine())
	m.Threads = float64(numThread)
	m.CGOCalls = float64(runtime.NumCgoCall())
	m.CPU = float64(runtime.NumCPU())

	m.DataPointReceived = DataPointReceived.get()
	m.DataPointAccepted = DataPointAccepted.get()
	m.DataPointRejectedBadRequest = DataPointRejectedBadRequest.get()
	m.DataPointDropped = DataPointDropped.get()
	m.SitesInCache = SitesInCache.get()
}

func (g *Sync) readGo(ts time.Time) {
	runtime.ReadMemStats(&g.ms)
	g.stats.From(&g.ms)
	g.stats.Timestamp = ts
}
