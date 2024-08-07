package sys

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/RoaringBitmap/roaring/v2/roaring64"
	"github.com/gernest/rows"
	"github.com/vinceanalytics/vince/internal/btx"
	"github.com/vinceanalytics/vince/internal/rbf"
	"github.com/vinceanalytics/vince/internal/rbf/cursor"
	chart "github.com/wcharczuk/go-chart/v2"
)

type chartBSI struct {
	ts         *roaring64.Bitmap
	ram        *roaring64.BSI
	histograms [3]*roaring64.BSI
	requests   *roaring64.BSI
}

func newChartBSI() *chartBSI {
	o := &chartBSI{
		ts:       roaring64.New(),
		ram:      roaring64.NewDefaultBSI(),
		requests: roaring64.NewDefaultBSI(),
	}
	o.histograms[0] = roaring64.NewDefaultBSI()
	o.histograms[1] = roaring64.NewDefaultBSI()
	o.histograms[2] = roaring64.NewDefaultBSI()
	return o
}

func (db *Store) Heap(w http.ResponseWriter, r *http.Request) error {
	b := roaring64.NewDefaultBSI()
	err := db.view(func(tx *rbf.Tx, shard uint64, f *rows.Row) error {
		return cursor.Tx(tx, "heap", func(c *rbf.Cursor) error {
			return btx.ExtractBSI(c, shard, f, func(column uint64, value int64) error {
				b.SetValue(column, value)
				return nil
			})
		})
	})
	if err != nil {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return err
	}
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
	b := roaring64.NewDefaultBSI()
	err := db.view(func(tx *rbf.Tx, shard uint64, f *rows.Row) error {
		return cursor.Tx(tx, "count", func(c *rbf.Cursor) error {
			return btx.ExtractBSI(c, shard, f, func(column uint64, value int64) error {
				b.SetValue(column, value)
				return nil
			})
		})
	})
	if err != nil {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return err
	}

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
	b0 := roaring64.NewDefaultBSI()
	b1 := roaring64.NewDefaultBSI()
	b2 := roaring64.NewDefaultBSI()
	err := db.view(func(tx *rbf.Tx, shard uint64, f *rows.Row) error {
		return errors.Join(
			cursor.Tx(tx, "b0", func(c *rbf.Cursor) error {
				return btx.ExtractBSI(c, shard, f, func(column uint64, value int64) error {
					b0.SetValue(column, value)
					return nil
				})
			}),
			cursor.Tx(tx, "b1", func(c *rbf.Cursor) error {
				return btx.ExtractBSI(c, shard, f, func(column uint64, value int64) error {
					b1.SetValue(column, value)
					return nil
				})
			}),
			cursor.Tx(tx, "b2", func(c *rbf.Cursor) error {
				return btx.ExtractBSI(c, shard, f, func(column uint64, value int64) error {
					b1.SetValue(column, value)
					return nil
				})
			}),
		)
	})
	if err != nil {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return err
	}

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
		XValues: make([]time.Time, 0, ex.GetCardinality()),
		YValues: make([]float64, 0, ex.GetCardinality()),
	}
	it := ex.Iterator()
	tEnd := ex.ReverseIterator().Next()
	vEnd, _ := b.GetValue(tEnd)
	tsEnd := time.UnixMilli(int64(tEnd)).UTC()
	for it.HasNext() {
		column := it.Next()
		ts := time.UnixMilli(int64(column)).UTC()
		data, _ := b.GetValue(column)
		dv := vEnd - data
		dt := tsEnd.Sub(ts)
		series.XValues = append(series.XValues, ts)
		series.YValues = append(series.YValues, float64(dv)/dt.Seconds())
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
