package dee

import (
	"bytes"
	"html/template"
	"sync"
	"time"

	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

var Map = template.FuncMap{
	"chart":       create,
	"title":       title,
	"size":        size,
	"style":       style,
	"class":       class,
	"padding":     padding,
	"strokeWidth": strokeWidth,
	"strokeColor": strokeColor,
	"dotColor":    dotColor,
	"dotWidth":    dotWidth,
	"hide":        hide,
	"axis":        axis,
	"render":      render,
	"series":      series,
}

func create() *chart.Chart {
	return &chart.Chart{
		YAxisSecondary: chart.HideYAxis(),
	}
}

func title(txt string, a *chart.Chart) *chart.Chart {
	a.Title = txt
	return a
}

func size(width, height int, a *chart.Chart) *chart.Chart {
	a.Width = width
	a.Height = height
	return a
}

func class(name string, o Style) *chart.Chart {
	o.x.ClassName = name
	return o.a
}

func padding(mode, pos string, px int, o Style) *chart.Chart {
	x := o.x
	switch pos {
	case "x":
		x.Padding.Left = px
		x.Padding.Right = px
	case "y":
		x.Padding.Top = px
		x.Padding.Bottom = px
	case "l":
		x.Padding.Left = px
	case "r":
		x.Padding.Right = px
	case "t":
		x.Padding.Top = px
	case "b":
		x.Padding.Bottom = px
	}
	return o.a
}

func strokeWidth(mode string, w float64, o Style) *chart.Chart {
	o.x.StrokeWidth = w
	return o.a
}

func strokeColor(mode string, c string, o Style) *chart.Chart {
	o.x.StrokeColor = drawing.ColorFromHex(c)
	return o.a
}

func dotColor(mode string, c string, o Style) *chart.Chart {
	o.x.DotColor = drawing.ColorFromHex(c)
	return o.a
}

func dotWidth(mode string, w float64, o Style) *chart.Chart {
	o.x.DotWidth = w
	return o.a
}

func hide(x *chart.Style, o Style) *chart.Chart {
	o.x.Hidden = true
	return o.a
}

type Style struct {
	a *chart.Chart
	x *chart.Style
}

func style(mode string, a *chart.Chart) (o Style) {
	o.a = a
	switch mode {
	case "title":
		o.x = &a.TitleStyle
	case "canvas":
		o.x = &a.Canvas
	case "bg":
		o.x = &a.Background
	case "x":
		o.x = &a.XAxis.Style
	case "xn":
		o.x = &a.XAxis.NameStyle
	case "xt":
		o.x = &a.XAxis.TickStyle
	case "xgj":
		o.x = &a.XAxis.GridMajorStyle
	case "xgi":
		o.x = &a.XAxis.GridMinorStyle
	case "y":
		o.x = &a.YAxis.Style
	case "yn":
		o.x = &a.YAxis.NameStyle
	case "yt":
		o.x = &a.YAxis.TickStyle
	case "ygj":
		o.x = &a.YAxis.GridMajorStyle
	case "ygi":
		o.x = &a.YAxis.GridMinorStyle
	default:
		o.x = &a.Background
	}
	return o
}

type Axis struct {
	a *chart.Chart
	x *chart.XAxis
	y *chart.YAxis
}

func axis(mode string, a *chart.Chart) Axis {
	switch mode {
	case "x":
		return Axis{a: a, x: &a.XAxis}
	case "y":
		return Axis{a: a, y: &a.YAxis}
	default:
		return Axis{a: a, x: &a.XAxis}
	}
}

func series(ts []time.Time, values []float64, a *chart.Chart) *chart.Chart {
	a.Series = append(a.Series, &chart.TimeSeries{
		XValues: ts,
		YValues: values,
	})
	return a
}

func render(a *chart.Chart) (template.HTML, error) {
	b := pool.Get().(*bytes.Buffer)
	defer pool.Put(b)
	b.Reset()
	err := a.Render(chart.SVG, b)
	if err != nil {
		return "", err
	}
	return template.HTML(b.String()), nil
}

var pool = &sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}
