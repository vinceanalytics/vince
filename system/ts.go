package system

import (
	"errors"
	"io"
	"os"
	"sort"
	"time"

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

func QueryStatsFile(path string, start, end time.Time, window time.Duration) (*StatsSeries, error) {
	ts, err := readStats(path, start.UnixMilli(), end.UnixMilli(), window.Milliseconds())
	if err != nil {
		return nil, err
	}
	shared := sharedTimestamp(start.UnixMilli(), end.UnixMilli())
	o := &StatsSeries{
		Timestamp: shared,
		Metrics:   make(map[string][]float64),
	}
	w := window.Milliseconds()
	for m, series := range ts.Metrics {
		switch m {
		case "counter":
			// use rate for counters
			o.Metrics[m] = roll(shared, ts.Timestamp, series, w, rate)
		case "gauge":
			// use sum for gauges
			o.Metrics[m] = roll(shared, ts.Timestamp, series, w, sum)
		}
	}
	return o, nil

}

func readStats(path string, start, end int64, window int64) (ts *StatsSeries, err error) {
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
		chunks := g.ColumnChunks()
		tsPage := chunks[0].Pages()
		var tsValues []int64
		var readAndExit bool
		for {
			page, err := tsPage.ReadPage()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
			}
			min, max, ok := page.Bounds()
			if !ok {
				tsPage.Close()
				// we are done return early
				if ts == nil {
					// we had read values before. Signal the end of reading
					return ts, nil
				}
				return nil, io.EOF
			}
			lower := min.Int64()
			if lower >= start {
				tsPage.Close()
				// we are done return early
				if ts == nil {
					// we had read values before. Signal the end of reading
					return ts, nil
				}
				return nil, io.EOF
			}
			tsValues = make([]int64, page.NumValues())
			page.Values().(parquet.Int64Reader).ReadInt64s(tsValues)

			upper := max.Int64()
			if end <= upper {
				n := sort.Search(len(tsValues), func(i int) bool {
					return tsValues[i] < end
				})
				tsValues = tsValues[:n]
				readAndExit = true
			}
		}
		tsPage.Close()
		for i, col := range chunks[1:] {
			pages := col.Pages()
			page, err := pages.ReadPage()
			if err != nil {
				if !errors.Is(err, io.EOF) {
					pages.Close()
					return nil, err
				}
			}
			name := schema[i+1].Name()
			o := make([]float64, len(tsValues))
			page.Values().(parquet.DoubleReader).ReadDoubles(o)
			ts.Metrics[name] = append(ts.Metrics[name], o...)
			pages.Close()
			if readAndExit {
				return ts, nil
			}
		}
	}
	return
}
