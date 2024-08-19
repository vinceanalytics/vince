package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
	"github.com/vinceanalytics/vince/internal/web/db"
	"github.com/vinceanalytics/vince/internal/web/db/plug"
	chart "github.com/wcharczuk/go-chart/v2"
)

func RequireSuper(h plug.Handler) plug.Handler {
	return func(db *db.Config, w http.ResponseWriter, r *http.Request) {
		if usr := db.CurrentUser(); usr != nil && usr.SuperUser {
			h(db, w, r)
			return
		}
		db.HTMLCode(http.StatusNotFound, w, e404, map[string]any{})
	}
}

func SystemHeap(db *db.Config, w http.ResponseWriter, r *http.Request) {
	b := db.Get().Heap()
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
		XAxis: chart.XAxis{
			ValueFormatter: chart.TimeMinuteValueFormatter,
		},
		YAxis: chart.YAxis{
			ValueFormatter: formatSize,
		},
		Series: []chart.Series{
			series,
		},
	}
	w.Header().Set("Content-Type", "image/png")
	err := graph.Render(chart.PNG, w)
	if err != nil {
		db.Logger().Error("serving heap graph", "err", err)
	}
}

func SystemData(db *db.Config, w http.ResponseWriter, r *http.Request) {
	b := db.Get().Size()
	ex := b.GetExistenceBitmap()
	series := chart.TimeSeries{
		Name:    "data_size",
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
		XAxis: chart.XAxis{
			ValueFormatter: chart.TimeMinuteValueFormatter,
		},
		YAxis: chart.YAxis{
			ValueFormatter: formatSize,
		},
		Series: []chart.Series{
			series,
		},
	}
	w.Header().Set("Content-Type", "image/png")
	err := graph.Render(chart.PNG, w)
	if err != nil {
		db.Logger().Error("serving heap graph", "err", err)
	}
}

func SystemMetrics(db *db.Config, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(db.Get().Sys())
}

func SystemRequests(db *db.Config, w http.ResponseWriter, r *http.Request) {
	b := db.Get().Requests()
	graph := chart.Chart{
		XAxis: chart.XAxis{
			ValueFormatter: chart.TimeMinuteValueFormatter,
		},
		Series: []chart.Series{rate(b, "request_per_second")},
	}
	w.Header().Set("Content-Type", "image/png")
	err := graph.Render(chart.PNG, w)
	if err != nil {
		db.Logger().Error("serving requests graph", "err", err)
	}
}

func SystemStats(db *db.Config, w http.ResponseWriter, r *http.Request) {
	db.HTML(w, system, map[string]any{
		"system": true,
	})
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
		series.XValues[i] = time.UnixMilli(int64(times[i]))
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
