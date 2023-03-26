package plot

import (
	"fmt"
	"strconv"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type axisOptions struct {
}

type Data struct {
	labels   []string
	datasets []DataSet
}

func (d *Data) prep() {
	if len(d.datasets) == 0 {
		d.datasets = []DataSet{
			{Values: make([]float64, len(d.labels))},
		}
	}
}

type DataSet struct {
	Name   string
	Values []float64
}

func newAxis() *axis {
	return &axis{
		chartType: "line",
		measures:  DEFAULT_MEASURE,
	}
}

type axisData struct {
	labels    []string
	positions []int
}

type axis struct {
	container, svg, svgDefs, titleEl, drawArea                                 *html.Node
	title, chartType                                                           string
	showTooltip, showLegend, isNavigable, animate, truncateLegends, continuous bool
	measures                                                                   Measure

	setMeasure                           func(*Measure)
	baseHeight, height, baseWidth, width int
	data                                 Data
	datasetLength, unitWidth, xOffset    int
	xAxis                                axisData
}

func (b *axis) config() {
	h := b.height
	if h == 0 {
		h = b.measures.BaseHeight
	}
	b.baseHeight = h
	b.height = h - b.measures.ExtraHeight()
}

func (b *axis) makeContainer() {
	b.container = &html.Node{
		Type: html.ElementNode,
		Data: atom.Div.String(),
	}
}

func (b *axis) calc() {
	b.calcXPositions()
	b.calcYAxisParameters(true)
}

func (b *axis) calcXPositions() {
	b.datasetLength = len(b.data.labels)
	b.unitWidth = b.width / b.datasetLength
	b.xOffset = b.unitWidth / 2
	b.xAxis.labels = b.data.labels
	b.xAxis.positions = make([]int, len(b.data.labels))
	for i := range b.xAxis.positions {
		b.xAxis.positions[i] = b.xOffset + i*b.unitWidth
	}
}

func (b *axis) calcYAxisParameters(withMinimum bool) {

}

func (b *axis) updateWidth() {
	b.width = b.baseWidth - b.measures.ExtraWidth()
}

func (b *axis) makeChartArea() {
	if b.svg != nil {
		b.container.RemoveChild(b.svg)
	}
	b.svg = makeSVGContainer(
		b.container, "frappe-chart chart",
		strconv.Itoa(b.baseWidth),
		strconv.Itoa(b.baseHeight),
	)
	b.svgDefs = makeSVGDefs(b.svg)
	m := b.measures
	if b.title != "" {
		b.titleEl = makeText(
			"title", m.Margin.Left, m.Margin.Top, b.title,
			textOptions{
				fontSize: m.TitleFontSize,
				fill:     "#666666",
				dy:       m.TitleFontSize,
			},
		)
	}
	top := m.TopOffset()
	b.drawArea = makeSVGGroup(
		b.chartType+"-chart chart-draw-area",
		fmt.Sprintf("translate(%d,%d)", m.LeftOffset(), top),
	)
	if b.title != "" {
		b.svg.AppendChild(b.titleEl)
	}
	b.svg.AppendChild(b.drawArea)
}

func (b *axis) draw() {
	b.updateWidth()
	b.makeChartArea()
}
