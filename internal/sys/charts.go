package sys

import (
	"fmt"
	"net/http"
	"time"

	"github.com/RoaringBitmap/roaring/v2/roaring64"
	chart "github.com/wcharczuk/go-chart/v2"
)

func (db *Store) Heap(w http.ResponseWriter, r *http.Request) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	b := &db.heap
	ex := b.GetExistenceBitmap()
	series := chart.TimeSeries{
		Name:    "heap",
		XValues: make([]time.Time, 0, ex.GetCardinality()),
		YValues: make([]float64, 0, ex.GetCardinality()),
	}
	it := ex.Iterator()
	for it.HasNext() {
		v := it.Next()
		series.XValues = append(series.XValues,
			time.UnixMilli(int64(v)).UTC())
		data, _ := b.GetValue(v)
		series.YValues = append(series.YValues, float64(data))
	}
	graph := chart.Chart{
		YAxis: chart.YAxis{
			ValueFormatter: formatSize,
		},
		Series: []chart.Series{
			series,
		},
	}
	w.Header().Set("Content-Type", "image/png")
	return graph.Render(chart.PNG, w)
}

// Renders rate of requests per second.
func (db *Store) Request(w http.ResponseWriter, r *http.Request) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	b := &db.requests
	graph := chart.Chart{
		YAxis: chart.YAxis{
			ValueFormatter: formatSize,
		},
		Series: []chart.Series{rate(b, "request_per_second")},
	}
	w.Header().Set("Content-Type", "image/png")
	return graph.Render(chart.PNG, w)
}

func (db *Store) Duration(w http.ResponseWriter, r *http.Request) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	b0 := &db.h0
	b1 := &db.h1
	b2 := &db.h2

	graph := chart.Chart{
		YAxis: chart.YAxis{
			ValueFormatter: formatSize,
		},
		Series: []chart.Series{
			rate(b0, "<= 0.5s"),
			rate(b1, "<= 1s"),
			rate(b2, ">  1s"),
		},
	}
	w.Header().Set("Content-Type", "image/png")
	graph.Elements = []chart.Renderable{
		chart.Legend(&graph),
	}
	return graph.Render(chart.PNG, w)
}

func rate(b *roaring64.BSI, name string) (series chart.TimeSeries) {
	ex := b.GetExistenceBitmap()
	series = chart.TimeSeries{
		Name:    name,
		XValues: make([]time.Time, ex.GetCardinality()),
		YValues: make([]float64, ex.GetCardinality()),
	}
	if ex.IsEmpty() {
		return
	}
	times := ex.ToArray()
	tEnd := times[len(times)-1]
	vEnd, _ := b.GetValue(tEnd)
	for i := range times {
		if i == 0 {
			// Assume the value didn't change
			series.XValues[i] = time.UnixMilli(int64(times[i]))
			continue
		}
		prevTs := times[i-1]
		prevValue, _ := b.GetValue(prevTs)
		dv := vEnd - prevValue
		dt := float64(tEnd-prevTs) / 1e3
		series.YValues[i] = float64(dv) / dt
	}
	return
}

func formatSize(v any) string {
	a, ok := v.(float64)
	if !ok {
		return fmt.Sprint(v)
	}
	if a <= (1 << 20) {
		return fmt.Sprintf("%.2fK", float64(a)/(1<<10))
	}
	if a <= (1 << 30) {
		return fmt.Sprintf("%.2fM", float64(a)/(1<<20))
	}
	return fmt.Sprintf("%.2fG", float64(a)/(1<<30))
}
