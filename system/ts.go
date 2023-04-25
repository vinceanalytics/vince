package system

import (
	"errors"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/segmentio/parquet-go"
)

type StatsSeries struct {
	Timestamp []int64
	Metrics   map[string][]float64
}

type Series struct {
	Name   string
	Values []float64
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
		if strings.HasSuffix(m, "_total") {
			// use rate for counters
			o.Metrics[m] = roll(shared, ts.Timestamp, series, w, rate)
		} else {
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
