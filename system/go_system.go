package system

import (
	"runtime"
	"time"
)

func (g *Sync) readGo(ts time.Time) {
	runtime.ReadMemStats(&g.ms)
	g.counters = g.counters[:0]
	g.gauges = g.gauges[:0]
	ms := &g.ms
	g.gauges = append(g.gauges, &Gauge{
		Timestamp: ts,
		Name:      "go_memstats_alloc_bytes",
		Value:     float64(ms.Alloc),
	})
	g.counters = append(g.counters, &Counter{
		Timestamp: ts,
		Name:      "go_memstats_alloc_bytes_total",
		Value:     float64(ms.TotalAlloc),
	})
	g.counters = append(g.counters, &Counter{
		Timestamp: ts,
		Name:      "go_memstats_frees_total",
		Value:     float64(ms.Frees),
	})
	g.gauges = append(g.gauges, &Gauge{
		Timestamp: ts,
		Name:      "go_memstats_gc_cpu_fraction",
		Value:     ms.GCCPUFraction,
	})
	g.gauges = append(g.gauges, &Gauge{
		Timestamp: ts,
		Name:      "go_memstats_gc_sys_bytes",
		Value:     float64(ms.GCSys),
	})
	g.gauges = append(g.gauges, &Gauge{
		Timestamp: ts,
		Name:      "go_memstats_heap_alloc_bytes",
		Value:     float64(ms.HeapAlloc),
	})
	g.gauges = append(g.gauges, &Gauge{
		Timestamp: ts,
		Name:      "go_memstats_heap_idle_bytes",
		Value:     float64(ms.HeapIdle),
	})
	g.gauges = append(g.gauges, &Gauge{
		Timestamp: ts,
		Name:      "go_memstats_heap_inuse_bytes",
		Value:     float64(ms.HeapInuse),
	})
	g.gauges = append(g.gauges, &Gauge{
		Timestamp: ts,
		Name:      "go_memstats_heap_objects",
		Value:     float64(ms.HeapObjects),
	})
	g.gauges = append(g.gauges, &Gauge{
		Timestamp: ts,
		Name:      "go_memstats_heap_objects",
		Value:     float64(ms.HeapReleased),
	})
	g.gauges = append(g.gauges, &Gauge{
		Timestamp: ts,
		Name:      "go_memstats_heap_sys_bytes",
		Value:     float64(ms.HeapSys),
	})
	g.counters = append(g.counters, &Counter{
		Timestamp: ts,
		Name:      "go_memstats_last_gc_time_seconds",
		Value:     float64(ms.LastGC) / 1e9,
	})
	g.gauges = append(g.gauges, &Gauge{
		Timestamp: ts,
		Name:      "go_memstats_lookups_total",
		Value:     float64(ms.Lookups),
	})
	g.counters = append(g.counters, &Counter{
		Timestamp: ts,
		Name:      "go_memstats_mallocs_total",
		Value:     float64(ms.Mallocs),
	})
	g.gauges = append(g.gauges, &Gauge{
		Timestamp: ts,
		Name:      "go_memstats_mcache_inuse_bytes",
		Value:     float64(ms.MCacheInuse),
	})
	g.gauges = append(g.gauges, &Gauge{
		Timestamp: ts,
		Name:      "go_memstats_mcache_sys_bytes",
		Value:     float64(ms.MCacheSys),
	})
	g.gauges = append(g.gauges, &Gauge{
		Timestamp: ts,
		Name:      "go_memstats_mspan_inuse_bytes",
		Value:     float64(ms.MSpanInuse),
	})
	g.gauges = append(g.gauges, &Gauge{
		Timestamp: ts,
		Name:      "go_memstats_mspan_sys_bytes",
		Value:     float64(ms.MSpanInuse),
	})
	g.gauges = append(g.gauges, &Gauge{
		Timestamp: ts,
		Name:      "go_memstats_next_gc_bytes",
		Value:     float64(ms.NextGC),
	})
	g.gauges = append(g.gauges, &Gauge{
		Timestamp: ts,
		Name:      "go_memstats_other_sys_bytes",
		Value:     float64(ms.OtherSys),
	})
	g.gauges = append(g.gauges, &Gauge{
		Timestamp: ts,
		Name:      "go_memstats_stack_inuse_bytes",
		Value:     float64(ms.StackInuse),
	})
	g.gauges = append(g.gauges, &Gauge{
		Timestamp: ts,
		Name:      "go_memstats_stack_sys_bytes",
		Value:     float64(ms.StackSys),
	})
	g.gauges = append(g.gauges, &Gauge{
		Timestamp: ts,
		Name:      "go_memstats_sys_bytes",
		Value:     float64(ms.Sys),
	})
	g.gauges = append(g.gauges, &Gauge{
		Timestamp: ts,
		Name:      "go_cgo_calls_count",
		Value:     float64(runtime.NumCgoCall()),
	})
	g.gauges = append(g.gauges, &Gauge{
		Timestamp: ts,
		Name:      "go_goroutines",
		Value:     float64(runtime.NumGoroutine()),
	})
	numThread, _ := runtime.ThreadCreateProfile(nil)
	g.gauges = append(g.gauges, &Gauge{
		Timestamp: ts,
		Name:      "go_goroutines",
		Value:     float64(numThread),
	})
}
