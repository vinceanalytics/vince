package plot

import "math"

const (
	LEGEND_ITEM_WIDTH               = 150
	SERIES_LABEL_SPACE_RATIO        = 0.6
	LINE_CHART_DOT_SIZE             = 4
	DOT_OVERLAY_SIZE_INCR           = 4
	PERCENTAGE_BAR_DEFAULT_HEIGHT   = 16
	HEATMAP_DISTRIBUTION_SIZE       = 5
	HEATMAP_SQUARE_SIZE             = 10
	HEATMAP_GUTTER_SIZE             = 2
	DEFAULT_CHAR_WIDTH              = 7
	TOOLTIP_POINTER_TRIANGLE_HEIGHT = 7.48
	ANGLE_RATIO                     = math.Pi / 180
	FULL_ANGLE                      = 360
)

var (
	DEFAULT_CHART_COLORS = []string{
		"pink",
		"blue",
		"green",
		"grey",
		"red",
		"yellow",
		"purple",
		"teal",
		"cyan",
		"orange",
	}
	HEATMAP_COLORS_GREEN = []string{
		"#ebedf0",
		"#c6e48b",
		"#7bc96f",
		"#239a3b",
		"#196127",
	}
	HEATMAP_COLORS_YELLOW = []string{
		"#ebedf0",
		"#fdf436",
		"#ffc700",
		"#ff9100",
		"#06001c",
	}
	DEFAULT_MEASURE = Measure{
		Margin: Dimension{
			Top:    10,
			Bottom: 10,
			Left:   20,
			Right:  20,
		},
		Padding: Dimension{
			Top:    20,
			Bottom: 40,
			Left:   30,
			Right:  10,
		},
		BaseHeight:    340,
		TitleHeight:   20,
		LegendHeight:  30,
		TitleFontSize: 12,
	}
)

type Measure struct {
	Margin, Padding                                      Dimension
	BaseHeight, TitleHeight, LegendHeight, TitleFontSize float64
}

type Dimension struct {
	Top, Bottom, Left, Right float64
}

func (m Measure) TopOffset() float64 {
	return m.TitleHeight + m.Margin.Top + m.Padding.Top
}

func (m Measure) LeftOffset() float64 {
	return m.Margin.Left + m.Padding.Left
}

func (m Measure) ExtraHeight() float64 {
	return m.Margin.Top + m.Margin.Bottom + m.Padding.Top + m.Padding.Bottom +
		m.TitleHeight + m.LegendHeight
}

func (m Measure) ExtraWidth() float64 {
	return m.Margin.Left + m.Margin.Right + m.Padding.Left + m.Padding.Right
}
