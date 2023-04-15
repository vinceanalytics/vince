package system

import (
	"runtime"
	"time"
)

type Stats struct {
	Timestamp     time.Time `parquet:"timestamp"`
	Alloc         uint64    `parquet:"go_memstats_alloc_bytes"`
	TotalAlloc    uint64    `parquet:"go_memstats_alloc_bytes_total"`
	Sys           uint64    `parquet:"go_memstats_sys_bytes"`
	Lookups       uint64    `parquet:"go_memstats_lookups_total"`
	Mallocs       uint64    `parquet:"go_memstats_mallocs_total"`
	Frees         uint64    `parquet:"go_memstats_frees_total"`
	HeapAlloc     uint64    `parquet:"go_memstats_heap_alloc_bytes"`
	HeapSys       uint64    `parquet:"go_memstats_heap_sys_bytes"`
	HeapIdle      uint64    `parquet:"go_memstats_heap_idle_bytes"`
	HeapInuse     uint64    `parquet:"go_memstats_heap_inuse_bytes"`
	HeapReleased  uint64    `parquet:"go_memstats_heap_released_bytes"`
	HeapObjects   uint64    `parquet:"go_memstats_heap_objects"`
	StackInuse    uint64    `parquet:"go_memstats_stack_inuse_bytes"`
	StackSys      uint64    `parquet:"go_memstats_stack_sys_bytes"`
	MSpanInuse    uint64    `parquet:"go_memstats_mspan_inuse_bytes"`
	MSpanSys      uint64    `parquet:"go_memstats_mspan_sys_bytes"`
	MCacheInuse   uint64    `parquet:"go_memstats_mcache_inuse_bytes"`
	MCacheSys     uint64    `parquet:"go_memstats_mcache_sys_bytes"`
	BuckHashSys   uint64    `parquet:"go_memstats_alloc_bytes_total"`
	GCSys         uint64    `parquet:"go_memstats_gc_sys_bytes"`
	OtherSys      uint64    `parquet:"go_memstats_other_sys_bytes"`
	NextGC        uint64    `parquet:"go_memstats_next_gc_bytes"`
	LastGC        float64   `parquet:"go_memstats_last_gc_time_seconds"`
	PauseTotalNs  float64   `parquet:"go_gc_duration_seconds_sum"`
	NumGC         uint32    `parquet:"go_gc_duration_seconds_count"`
	NumForcedGC   uint32    `parquet:"go_gc_forced_count"`
	GCCPUFraction float64   `parquet:"go_memstats_gc_cpu_fraction"`
	CGOCalls      int64     `parquet:"go_cgo_calls_count"`
	CPU           int       `parquet:"go_cpu_count"`
	Goroutines    int       `parquet:"go_goroutines"`
	MaxProcs      int       `parquet:"go_gomaxprocs"`
	Threads       int       `parquet:"go_threads"`

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
	m.Alloc = o.Alloc
	m.TotalAlloc = o.TotalAlloc
	m.Sys = o.Sys
	m.Lookups = o.Lookups
	m.Mallocs = o.Mallocs
	m.Frees = o.Frees
	m.HeapAlloc = o.HeapAlloc
	m.HeapSys = o.HeapSys
	m.HeapIdle = o.HeapIdle
	m.HeapInuse = o.HeapInuse
	m.HeapReleased = o.HeapReleased
	m.HeapObjects = o.HeapObjects
	m.StackInuse = o.StackInuse
	m.StackSys = o.StackSys
	m.MSpanInuse = o.MSpanInuse
	m.MSpanSys = o.MSpanSys
	m.MCacheInuse = o.MCacheInuse
	m.MCacheSys = o.MCacheSys
	m.BuckHashSys = o.BuckHashSys
	m.GCSys = o.GCSys
	m.OtherSys = o.OtherSys
	m.NextGC = o.NextGC
	m.LastGC = float64(o.LastGC) / 1e9
	m.PauseTotalNs = float64(o.PauseTotalNs) / 1e9
	m.NumGC = o.NumGC
	m.NumForcedGC = o.NumForcedGC
	m.GCCPUFraction = o.GCCPUFraction
	m.MaxProcs = runtime.GOMAXPROCS(0)
	m.Goroutines = runtime.NumGoroutine()
	m.Threads = numThread
	m.CGOCalls = runtime.NumCgoCall()
	m.CPU = runtime.NumCPU()

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
