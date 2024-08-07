package sys

import (
	"bytes"
	"os"
	"testing"
	"time"

	chart "github.com/wcharczuk/go-chart/v2"
)

func TestY(t *testing.T) {

	graph := chart.Chart{
		Series: []chart.Series{
			chart.TimeSeries{
				Name: "heap",
				XValues: []time.Time{
					time.Now().AddDate(0, 0, -10),
					time.Now().AddDate(0, 0, -9),
					time.Now().AddDate(0, 0, -8),
					time.Now().AddDate(0, 0, -7),
					time.Now().AddDate(0, 0, -6),
					time.Now().AddDate(0, 0, -5),
					time.Now().AddDate(0, 0, -4),
					time.Now().AddDate(0, 0, -3),
					time.Now().AddDate(0, 0, -2),
					time.Now().AddDate(0, 0, -1),
					time.Now(),
				},
				YValues: []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0, 11.0},
			},
			chart.TimeSeries{
				Name: "total_requests",
				XValues: []time.Time{
					time.Now().AddDate(0, 0, -10),
					time.Now().AddDate(0, 0, -9),
					time.Now().AddDate(0, 0, -8),
					time.Now().AddDate(0, 0, -7),
					time.Now().AddDate(0, 0, -6),
					time.Now().AddDate(0, 0, -5),
					time.Now().AddDate(0, 0, -4),
					time.Now().AddDate(0, 0, -3),
					time.Now().AddDate(0, 0, -2),
					time.Now().AddDate(0, 0, -1),
					time.Now(),
				},
				YValues: []float64{1.0, 2.0, 3.0, 4.0, 2.0, 6.0, 7.0, 8.0, 9.0, 10.0, 11.0},
			},
		},
	}
	graph.Elements = []chart.Renderable{
		chart.Legend(&graph),
	}
	var o bytes.Buffer
	graph.Render(chart.PNG, &o)
	os.WriteFile("ts.png", o.Bytes(), 0600)

}
