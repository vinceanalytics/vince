package graph

import (
	"fmt"
	"io"

	"github.com/wcharczuk/go-chart/v2"
)

func Trend(width, height int, series []float64, w io.Writer) error {
	x := make([]float64, len(series))
	for i := range x {
		x[i] = float64(i)
	}
	anno := make([]chart.Value2, len(series))
	for i := range anno {
		anno[i] = chart.Value2{
			XValue: float64(i),
			YValue: series[i],
			Label:  fmt.Sprint(series[i]),
		}
		x[i] = float64(i)
	}
	graph := chart.Chart{
		Width:  width,
		Height: height,
		XAxis:  chart.HideXAxis(),
		YAxis:  chart.HideYAxis(),
		Series: []chart.Series{
			chart.ContinuousSeries{
				Style: chart.Style{
					StrokeColor: TrendStroke,
					FillColor:   TrendFill,
				},
				XValues: x,
				YValues: series,
			},
			chart.AnnotationSeries{
				Annotations: anno,
			},
		},
	}
	return graph.Render(chart.SVG, w)
}
