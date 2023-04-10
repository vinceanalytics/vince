package graph

import (
	"bytes"
	"html/template"
	"io"
	"strings"

	"github.com/wcharczuk/go-chart/v2"
)

func Trend(width, height int, series []float64, w io.Writer) error {
	x := make([]float64, len(series))
	for i := range x {
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
		},
	}
	return graph.Render(chart.SVG, w)
}

func SiteTrend() template.HTML {
	var b bytes.Buffer
	err := Trend(1613, 240, []float64{1.0, 10.0, 8.0, 4.0, 5.0}, &b)
	if err != nil {
		return ""
	}
	s := strings.Replace(b.String(), fix, replace, 1)
	return template.HTML(s)
}

const fix = `style="stroke-width:1;stroke:rgba(51,51,51,1.0);fill:none"`
const replace = `style="display:none;"`
