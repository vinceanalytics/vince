package system

import (
	"errors"
	"io"
	"os"

	"github.com/segmentio/parquet-go"
)

type StatsSeries struct {
	Timestamp []int64

	Metrics map[string][]float64
}

type Series struct {
	Name   string
	Values []float64
}

// MetricType maps stats to stats type  counter or gauge
var MetricType = map[string]string{
	"data_point_accepted_total":             "counter",
	"data_point_dropped_total":              "counter",
	"data_point_received_total":             "counter",
	"data_point_rejected_bad_request_total": "counter",
	"go_cgo_calls_count":                    "counter",
	"go_cpu_count":                          "counter",
	"go_gc_duration_seconds_count":          "counter",
	"go_gc_duration_seconds_sum":            "counter",
	"go_gc_forced_count":                    "counter",
	"go_gomaxprocs":                         "counter",
	"go_goroutines":                         "counter",
	"go_memstats_alloc_bytes":               "counter",
	"go_memstats_alloc_bytes_total":         "counter",
	"go_memstats_frees_total":               "counter",
	"go_memstats_gc_cpu_fraction":           "counter",
	"go_memstats_gc_sys_bytes":              "counter",
	"go_memstats_heap_alloc_bytes":          "counter",
	"go_memstats_heap_idle_bytes":           "counter",
	"go_memstats_heap_inuse_bytes":          "counter",
	"go_memstats_heap_objects":              "counter",
	"go_memstats_heap_released_bytes":       "counter",
	"go_memstats_heap_sys_bytes":            "counter",
	"go_memstats_last_gc_time_seconds":      "counter",
	"go_memstats_lookups_total":             "counter",
	"go_memstats_mallocs_total":             "counter",
	"go_memstats_mcache_inuse_bytes":        "counter",
	"go_memstats_mcache_sys_bytes":          "counter",
	"go_memstats_mspan_inuse_bytes":         "counter",
	"go_memstats_mspan_sys_bytes":           "counter",
	"go_memstats_next_gc_bytes":             "counter",
	"go_memstats_other_sys_bytes":           "counter",
	"go_memstats_stack_inuse_bytes":         "counter",
	"go_memstats_stack_sys_bytes":           "counter",
	"go_memstats_sys_bytes":                 "counter",
	"go_threads":                            "counter",
	"sites_in_cache":                        "gauge",
}

func readStats(path string) (ts *StatsSeries, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	r, err := parquet.OpenFile(f, stat.Size())
	if err != nil {
		return nil, err
	}
	ts = &StatsSeries{
		Metrics: make(map[string][]float64),
	}
	schema := r.Schema().Fields()

	for _, g := range r.RowGroups() {
		for i, col := range g.ColumnChunks() {
			pages := col.Pages()
			for {
				page, err := pages.ReadPage()
				if err != nil {
					if errors.Is(err, io.EOF) {
						break
					}
				}
				if i == 0 {
					// first column is for timestamp
					o := make([]int64, page.NumValues())
					page.Values().(parquet.Int64Reader).ReadInt64s(o)
					ts.Timestamp = append(ts.Timestamp, o...)
				} else {
					name := schema[i].Name()
					o := make([]float64, page.NumValues())
					page.Values().(parquet.DoubleReader).ReadDoubles(o)
					ts.Metrics[name] = append(ts.Metrics[name], o...)
				}
			}
			pages.Close()
		}
	}
	return
}
