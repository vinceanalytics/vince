package graph

import (
	"io"

	"github.com/wcharczuk/go-chart"
)

func Trend(width, height int, series []float64, w io.Writer) error {
	x := make([]float64, len(series))
	for i := range x {
		x[i] = float64(i)
	}
	graph := chart.Chart{
		Width:  width,
		Height: height,
		Series: []chart.Series{
			chart.ContinuousSeries{
				Style: chart.Style{
					Show:        true,
					StrokeColor: TrendStroke,
					FillColor:   TrendFill,
				},
				XValues: x,
				YValues: []float64{1.0, 10.0, 8.0, 4.0, 5.0},
			},
		},
	}
	return graph.Render(chart.SVG, w)
}
