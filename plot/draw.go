package plot

import (
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

const (
	AXIS_TICK_LENGTH = 6
	LABEL_MARGIN     = 4
	LABEL_WIDTH      = 25
	TOTAL_PADDING    = 120
	LABEL_MAX_CHARS  = 18
	FONT_SIZE        = 10
	BASE_LINE_COLOR  = "#E2E6E9"
)

type createOptions struct {
	inside, around *html.Node
	style          map[string]string
	innerHtml      string
	attr           []html.Attribute
}

func createSVG(tag string, o createOptions) *html.Node {
	e := &html.Node{
		Type:      html.ElementNode,
		Namespace: "http://www.w3.org/2000/svg",
		Data:      tag,
	}
	if o.inside != nil {
		o.inside.AppendChild(e)
	}
	if o.around != nil {
		o.around.Parent.InsertBefore(e, o.around)
		e.AppendChild(o.around)
	}
	if len(o.style) > 0 {
		e.Attr = append(e.Attr, style(o.style))
	}
	if o.innerHtml != "" {
		e.Data = o.innerHtml
	}
	if len(o.attr) > 0 {
		e.Attr = append(e.Attr, o.attr...)
	}
	return e
}

func style(m map[string]string) html.Attribute {
	// make sure this is idempotent. Sort keys used for the attribute
	ls := make([]string, 0, len(m))
	for k := range m {
		ls = append(ls, k)
	}
	sort.Strings(ls)
	var s strings.Builder
	for i := range ls {
		s.WriteString(ls[i])
		s.WriteByte(':')
		s.WriteString(m[ls[i]])
		s.WriteByte(';')
	}
	return html.Attribute{
		Key: "style",
		Val: s.String(),
	}
}

func renderVerticalGradient(svgDefElem *html.Node, gradientId string) *html.Node {
	return createSVG("linearGradient", createOptions{
		inside: svgDefElem,
		attr: []html.Attribute{
			{Key: "id", Val: gradientId},
			{Key: "x1", Val: "0"},
			{Key: "x2", Val: "0"},
			{Key: "y1", Val: "0"},
			{Key: "y2", Val: "1"},
		},
	})
}

func setGradientStop(gradElem *html.Node, offset, color, opacity string) *html.Node {
	return createSVG("stop", createOptions{
		inside: gradElem,
		style: map[string]string{
			"stop-color": color,
		},
		attr: []html.Attribute{
			{Key: "offset", Val: offset},
			{Key: "stop-opacity", Val: opacity},
		},
	})
}

func makeSVGContainer(parent *html.Node, className, width, height string) *html.Node {
	return createSVG("svg", createOptions{
		inside: parent,
		attr: []html.Attribute{
			{Key: "class", Val: className},
			{Key: "width", Val: width},
			{Key: "height", Val: height},
		},
	})
}

func makeSVGDefs(svgContainer *html.Node) *html.Node {
	return createSVG("defs", createOptions{
		inside: svgContainer,
	})
}

func makeSVGGroup(className, transform string, parent ...*html.Node) *html.Node {
	o := createOptions{
		attr: []html.Attribute{
			{Key: "class", Val: className},
		},
	}
	if len(parent) > 0 {
		o.inside = parent[0]
	}
	o.attr = append(o.attr, html.Attribute{Key: "transform", Val: transform})
	return createSVG("g", o)
}

func makePath(pathStr, className, stroke, fill, strokeWidth string) *html.Node {
	if stroke == "" {
		stroke = "none"
	}
	if fill == "" {
		fill = "none"
	}
	if strokeWidth == "" {
		strokeWidth = "2"
	}
	return createSVG("path", createOptions{
		attr: []html.Attribute{
			{Key: "class", Val: className},
			{Key: "d", Val: pathStr},
		},
		style: map[string]string{
			"stroke":       stroke,
			"fill":         fill,
			"stroke-width": strokeWidth,
		},
	})
}

func makeGradient(svgDefElem *html.Node, color string, lighter bool) string {
	gradientId :=
		"path-fill-gradient" + "-" + color + "-"
	if lighter {
		gradientId += "lighter"
	} else {
		gradientId += "default"
	}
	gradientDef := renderVerticalGradient(svgDefElem, gradientId)
	opacities := []string{"1", "0.6", "0.2"}
	if lighter {
		opacities = []string{"0.4", "0.2", "0"}
	}
	setGradientStop(gradientDef, "0%", color, opacities[0])
	setGradientStop(gradientDef, "50%", color, opacities[1])
	setGradientStop(gradientDef, "100%", color, opacities[2])
	return gradientId
}

func rightRoundedBar(x, width, height float64) string {
	radius := height / 2
	xOffset := width - radius
	return fmt.Sprintf("M%d,0 h%d q%d,0 %d,%d q0,%d -%d,%d h-%d v%d",
		x, xOffset, radius, radius, radius, radius, radius, radius, xOffset, height)
}

func leftRoundedBar(x, width, height float64) string {
	radius := height / 2
	xOffset := width - radius
	return fmt.Sprintf("M%d,0 h%d v%d h-%d q-%d, 0 -%d,-%d q0,-%d %d,-%dz",
		x+radius, xOffset, height, xOffset, radius, radius, radius, radius, radius, radius)
}

func percentageBar(x, y, width, height float64, isFirst, isLast bool, fill string) *html.Node {
	if fill == "" {
		fill = "none"
	}
	if isLast {
		pathStr := rightRoundedBar(x, width, height)
		return makePath(pathStr, "percentage-bar", "", fill, "")
	}
	if isFirst {
		pathStr := leftRoundedBar(x, width, height)
		return makePath(pathStr, "percentage-bar", "", fill, "")
	}
	return createSVG("rect", createOptions{
		attr: []html.Attribute{
			{Key: "class", Val: "percentage-bar"},
			{Key: "x", Val: formatFloat(x)},
			{Key: "y", Val: formatFloat(y)},
			{Key: "width", Val: formatFloat(width)},
			{Key: "height", Val: formatFloat(height)},
			{Key: "fill", Val: fill},
		},
	})
}

func heatSquare(className string, x, y, size, radius float64, fill string, data ...html.Attribute) *html.Node {
	if fill == "" {
		fill = "none"
	}
	o := createOptions{
		attr: []html.Attribute{
			{Key: "class", Val: className},
			{Key: "x", Val: formatFloat(x)},
			{Key: "y", Val: formatFloat(y)},
			{Key: "width", Val: formatFloat(size)},
			{Key: "height", Val: formatFloat(size)},
			{Key: "rx", Val: formatFloat(radius)},
			{Key: "fill", Val: fill},
		},
	}
	o.attr = append(o.attr, data...)
	return createSVG("rect", o)
}

func legendDot(x, y, size, radius float64, fill, label, value string, fontSize float64, truncate bool) *html.Node {
	if fill == "" {
		fill = "none"
	}
	if truncate {
		label = truncateString(label, LABEL_MAX_CHARS)
	}
	if fontSize == 0 {
		fontSize = FONT_SIZE
	}
	o := createOptions{
		attr: []html.Attribute{
			{Key: "class", Val: "legend-dot"},
			{Key: "x", Val: "0"},
			{Key: "y", Val: formatFloat(4 - size)},
			{Key: "height", Val: formatFloat(size)},
			{Key: "width", Val: formatFloat(size)},
			{Key: "rx", Val: formatFloat(radius)},
			{Key: "fill", Val: fill},
		},
	}

	textLabel := createSVG("text", createOptions{
		attr: []html.Attribute{
			{Key: "class", Val: "legend-dataset-label"},
			{Key: "y", Val: "0"},
			{Key: "x", Val: formatFloat(size)},
			{Key: "dx", Val: formatFloat(fontSize) + "px"},
			{Key: "dy", Val: formatFloat(fontSize/3) + "px"},
			{Key: "font-size", Val: strconv.FormatFloat(float64(fontSize)*1.6, 'f', -1, 64) + "px"},
			{Key: "text-anchor", Val: "start"},
		},
		innerHtml: label,
	})
	var textValue *html.Node
	if value != "" {
		textValue = createSVG("text", createOptions{
			attr: []html.Attribute{
				{Key: "class", Val: "legend-dataset-value"},
				{Key: "x", Val: formatFloat(size)},
				{Key: "y", Val: formatFloat(FONT_SIZE + 10)},
				{Key: "dx", Val: formatFloat(FONT_SIZE) + "px"},
				{Key: "dy", Val: formatFloat(FONT_SIZE/3) + "px"},
				{Key: "font-size", Val: strconv.FormatFloat(float64(fontSize)*1.2, 'f', -1, 64) + "px"},
				{Key: "text-anchor", Val: "start"},
			},
			innerHtml: value,
		})
	}
	group := createSVG("g", createOptions{
		attr: []html.Attribute{
			{Key: "transform", Val: fmt.Sprintf("translate(%f, %f)", x, y)},
		},
	})
	group.AppendChild(createSVG("rect", o))
	group.AppendChild(textLabel)
	if textValue != nil {
		group.AppendChild(textLabel)
	}
	return group
}

func truncateString(txt string, n int) string {
	if txt == "" {
		return ""
	}
	if len(txt) > n {
		return txt[:n-3] + "..."
	}
	return txt
}

type textOptions struct {
	fontSize, dy     float64
	fill, textAnchor string
}

func makeText(className string, x, y float64, content string, o textOptions) *html.Node {
	if o.fontSize == 0 {
		o.fontSize = FONT_SIZE
	}
	if o.dy == 0 {
		o.dy = o.fontSize / 2
	}
	if o.fill == "" {
		o.fill = "var(--charts-label-color)"
	}
	if o.textAnchor == "" {
		o.textAnchor = "start"
	}
	return createSVG("text", createOptions{
		innerHtml: content,
		attr: []html.Attribute{
			{Key: "x", Val: formatFloat(x)},
			{Key: "y", Val: formatFloat(y)},
			{Key: "dy", Val: formatFloat(o.dy) + "px"},
			{Key: "font-size", Val: formatFloat(o.fontSize) + "px"},
			{Key: "fill", Val: o.fill},
			{Key: "text-anchor", Val: o.textAnchor},
		},
	})
}

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

type verLineOptions struct {
	className, lineType, stroke string
}

func makeVertLine(x float64, label string, y1, y2 float64, o verLineOptions) *html.Node {
	if o.stroke == "" {
		o.stroke = BASE_LINE_COLOR
	}
	l := createSVG("line", createOptions{
		style: map[string]string{
			"stroke": o.stroke,
		},
		attr: []html.Attribute{
			{Key: "class", Val: "line-vertical " + o.className},
			{Key: "x1", Val: "0"},
			{Key: "x2", Val: "0"},
			{Key: "y1", Val: formatFloat(y1)},
			{Key: "y2", Val: formatFloat(y2)},
		},
	})
	y := y1 - LABEL_MARGIN - FONT_SIZE
	if y1 > y2 {
		y = y1 + LABEL_MARGIN
	}
	text := createSVG("text", createOptions{
		innerHtml: label,
		attr: []html.Attribute{
			{Key: "x", Val: "0"},
			{Key: "y", Val: formatFloat(y)},
			{Key: "dy", Val: formatFloat(FONT_SIZE) + "px"},
			{Key: "font-size", Val: formatFloat(FONT_SIZE) + "px"},
			{Key: "text-anchor", Val: "middle"},
		},
	})
	line := createSVG("g", createOptions{
		attr: []html.Attribute{
			{Key: "transform", Val: fmt.Sprintf("translate(%f, 0)", x)},
		},
	})
	line.AppendChild(l)
	line.AppendChild(text)
	return line
}

type horiLineOptions struct {
	className, title, stroke, lineType, alignment string
	shortenNumbers                                bool
}

func makeHoriLine(y float64, label any, x1, x2 float64, o horiLineOptions) *html.Node {
	if o.stroke == "" {
		o.stroke = BASE_LINE_COLOR
	}
	if o.alignment == "" {
		o.alignment = "left"
	}
	if o.shortenNumbers {
		label = shortenLargeNumber(label)
	}
	className := "line-horizontal " + o.className + o.lineType
	var textXPos float64
	if o.alignment == "left" {
		if o.title != "" {
			textXPos = x1 - LABEL_MARGIN + LABEL_WIDTH
		} else {
			textXPos = x1 - LABEL_MARGIN
		}
	} else {
		if o.title != "" {
			textXPos = x2 + LABEL_MARGIN*4 - LABEL_WIDTH
		} else {
			textXPos = x2 + LABEL_MARGIN*4
		}
	}
	lineX1Post := x1
	if o.title != "" {
		lineX1Post = x1 + LABEL_WIDTH
	}
	lineX2Post := x2
	if o.title != "" {
		lineX2Post = x2 - LABEL_WIDTH
	}
	l := createSVG("line", createOptions{
		attr: []html.Attribute{
			{Key: "class", Val: className},
			{Key: "x1", Val: formatFloat(lineX1Post)},
			{Key: "x2", Val: formatFloat(lineX2Post)},
			{Key: "y1", Val: "0"},
			{Key: "y2", Val: "0"},
		},
		style: map[string]string{
			"stroke": o.stroke,
		},
	})
	a := "start"
	if x1 < x2 {
		a = "end"
	}
	text := createSVG("text", createOptions{
		attr: []html.Attribute{
			{Key: "x", Val: formatFloat(textXPos)},
			{Key: "y", Val: "0"},
			{Key: "dy", Val: formatFloat(FONT_SIZE/2-2) + "px"},
			{Key: "font-size", Val: formatFloat(FONT_SIZE) + "px"},
			{Key: "text-anchor", Val: a},
		},
		innerHtml: fmt.Sprint(label),
	})
	line := createSVG("g", createOptions{
		attr: []html.Attribute{
			{Key: "transform", Val: fmt.Sprintf("translate(0, %d)", y)},
			{Key: "stroke-opacity", Val: "1"},
		},
	})
	if text.Data == "0" {
		line.Attr = append(line.Attr, style(map[string]string{
			"stroke": "rgba(27, 31, 35, 0.6)",
		}))
	}
	line.AppendChild(l)
	line.AppendChild(text)
	return line
}
func shortenLargeNumber(a any) string {
	var n float64
	switch e := a.(type) {
	case string:
		if e == "" {
			return ""
		}
		n, _ = strconv.ParseFloat(e, 64)
	case int:
		n = float64(e)
	case float64:
		n = float64(e)
	}
	p := math.Floor(math.Log10(math.Abs(n)))
	if p < 2 {
		return strconv.FormatFloat(n, 'f', -1, 64)
	}
	l := math.Floor(p / 3)
	s := math.Pow(10, p-l*3) * +(n / math.Pow(10, p))
	s = math.Round(s*100) / 100
	return strconv.FormatFloat(s, 'f', -1, 64) + []string{"", "K", "M", "B", "T"}[int(l)]
}

type axisLabelOptions struct {
	title, position string
	height, width   float64
}

func generateAxisLabel(o axisLabelOptions) *html.Node {
	if o.title == "" {
		return nil
	}
	var y float64
	if o.position == "left" {
		y = (o.height-TOTAL_PADDING)/2 +
			float64(len(o.title)*5)/2
	} else {
		y = (o.height-TOTAL_PADDING)/2 -
			float64(len(o.title)*5)/2
	}
	var x float64
	if o.position != "left" {
		x = o.width
	}
	var y2 float64 = FONT_SIZE + LABEL_WIDTH*-1
	if o.position == "left" {
		y2 = FONT_SIZE - LABEL_WIDTH
	}
	rotation := "rotate(270)"
	if o.position == "right" {
		rotation = "rotate(90)"
	}
	labelSvg := createSVG("text", createOptions{
		innerHtml: o.title,
		attr: []html.Attribute{
			{Key: "class", Val: "chart-label"},
			{Key: "x", Val: "0"},
			{Key: "y", Val: "0"},
			{Key: "dy", Val: formatFloat(y2) + "px"},
			{Key: "font-size", Val: formatFloat(FONT_SIZE) + "px"},
			{Key: "text-anchor", Val: "start"},
		},
	})
	wrapper := createSVG("g", createOptions{
		attr: []html.Attribute{
			{Key: "x", Val: "0"},
			{Key: "y", Val: "0"},
			{Key: "transformBox", Val: "fill-box"},
			{Key: "transform", Val: fmt.Sprintf("translate(%f, %f) %s", x, y, rotation)},
			{Key: "class", Val: "test-" + o.position},
		},
	})
	wrapper.AppendChild(labelSvg)
	return wrapper
}

type yLineOptions struct {
	className, title, lineType, mode, pos, stroke string
	offset                                        float64
	shortenNumbers                                bool
}

func yLine(y float64, label any, width float64, o yLineOptions) *html.Node {
	if o.pos == "" {
		o.pos = "left"
	}
	if o.mode == "" {
		o.mode = "span"
	}
	if o.stroke == "" {
		o.stroke = BASE_LINE_COLOR
	}
	var x1 float64 = -1 * AXIS_TICK_LENGTH
	var x2 float64
	if o.mode == "span" {
		x2 = width + AXIS_TICK_LENGTH
	}
	if o.mode == "tick" && o.pos == "right" {
		x1 = width + AXIS_TICK_LENGTH
		x2 = width
	}
	x1 += o.offset
	x2 += o.offset
	if v, ok := label.(float64); ok {
		label = math.Round(v)
	}
	return makeHoriLine(y, label, x1, x2, horiLineOptions{
		stroke:         o.stroke,
		className:      o.className,
		lineType:       o.lineType,
		alignment:      o.pos,
		title:          o.title,
		shortenNumbers: o.shortenNumbers,
	})
}

func xLine(x float64, label string, height float64, o yLineOptions) *html.Node {
	if o.pos == "" {
		o.pos = "bottom"
	}
	if o.mode == "" {
		o.mode = "span"
	}
	if o.stroke == "" {
		o.stroke = BASE_LINE_COLOR
	}
	y1 := height + AXIS_TICK_LENGTH
	y2 := height
	if o.mode == "span" {
		y2 = -1 * AXIS_TICK_LENGTH
	}
	if o.mode == "tick" && o.pos == "top" {
		y1 = -1 * AXIS_TICK_LENGTH
		y2 = 0
	}
	return makeVertLine(x, label, y1, y2, verLineOptions{
		stroke:    o.stroke,
		className: o.className,
		lineType:  o.lineType,
	})
}

type yMarkerOptions struct {
	className, pos, stroke, lineType string
}

func yMarker(y float64, label string, width float64, o yLineOptions) *html.Node {
	if o.pos == "" {
		o.pos = "right"
	}
	if o.lineType == "" {
		o.lineType = "dashed"
	}
	x := width - float64(len(label)*5) - LABEL_MARGIN
	if o.pos == "left" {
		x = LABEL_MARGIN
	}
	if o.stroke == "" {
		o.stroke = BASE_LINE_COLOR
	}
	labelSvg := createSVG("text", createOptions{
		innerHtml: label,
		attr: []html.Attribute{
			{Key: "x", Val: formatFloat(x)},
			{Key: "y", Val: "0"},
			{Key: "dy", Val: formatFloat(FONT_SIZE/-2) + "px"},
			{Key: "font-size", Val: formatFloat(FONT_SIZE) + "px"},
			{Key: "text-anchor", Val: "start"},
		},
	})
	line := makeHoriLine(y, "", 0, width, horiLineOptions{
		stroke:    o.stroke,
		className: o.className,
		lineType:  o.lineType,
	})
	line.AppendChild(labelSvg)
	return line
}

type yRegionOptions struct {
	stroke, fill, pos string
}

func yRegion(y1, y2, width float64, label string, o yRegionOptions) *html.Node {
	if o.stroke == "" {
		o.stroke = BASE_LINE_COLOR
	}
	if o.fill == "" {
		o.fill = "rgba(228, 234, 239, 0.49)"
	}
	height := y1 - y2
	rect := createSVG("rect", createOptions{
		style: map[string]string{
			"fill":             o.fill,
			"stroke":           o.stroke,
			"stroke-dasharray": fmt.Sprintf("%d, %d", width, height),
		},
		attr: []html.Attribute{
			{Key: "x", Val: "0"},
			{Key: "y", Val: "0"},
			{Key: "width", Val: formatFloat(width)},
			{Key: "height", Val: formatFloat(height)},
		},
	})
	if o.pos == "" {
		o.pos = "right"
	}
	x := width - float64(len(label)*4) - LABEL_MARGIN
	if o.pos == "left" {
		x = LABEL_MARGIN
	}
	labelSvg := createSVG("text", createOptions{
		innerHtml: label,
		attr: []html.Attribute{
			{Key: "class", Val: "chart-label"},
			{Key: "x", Val: formatFloat(x)},
			{Key: "y", Val: "0"},
			{Key: "dy", Val: formatFloat(FONT_SIZE/-2) + "px"},
			{Key: "font-size", Val: formatFloat(FONT_SIZE) + "px"},
			{Key: "text-anchor", Val: "start"},
		},
	})
	region := createSVG("g", createOptions{
		attr: []html.Attribute{
			{Key: "transform", Val: fmt.Sprintf("translate(0, %d)", y2)},
		},
	})
	region.AppendChild(rect)
	region.AppendChild(labelSvg)
	return region
}

func getBarHeightAndYAttr(yTop, zeroLine float64) (height, y float64) {
	if yTop <= zeroLine {
		height = zeroLine - yTop
		y = yTop
	} else {
		height = yTop - zeroLine
		y = zeroLine
	}
	return
}

type datasetMeta struct {
	zeroLine, minHeight float64
}

func datasetBar(x, yTop, width float64, color, label string, index int, offset float64, m datasetMeta) *html.Node {
	height, y := getBarHeightAndYAttr(yTop, m.zeroLine)
	y -= offset
	if height == 0 {
		height = m.minHeight
		y -= m.minHeight
	}
	rect := createSVG("rect", createOptions{
		style: map[string]string{
			"fill": color,
		},
		attr: []html.Attribute{
			{Key: "class", Val: "bar mini"},
			{Key: "bar mini", Val: strconv.Itoa(index)},
			{Key: "x", Val: formatFloat(x)},
			{Key: "y", Val: formatFloat(y)},
			{Key: "width", Val: formatFloat(width)},
			{Key: "height", Val: formatFloat(height)},
		},
	})
	if label == "" {
		return rect
	}
	for i := range rect.Attr {
		if rect.Attr[i].Key == "x" || rect.Attr[i].Key == "y" {
			rect.Attr[i].Val = "0"
		}
	}
	text := createSVG("text", createOptions{
		innerHtml: label,
		attr: []html.Attribute{
			{Key: "class", Val: "data-point-value"},
			{Key: "x", Val: formatFloat(width / 2)},
			{Key: "y", Val: "0"},
			{Key: "dy", Val: formatFloat((FONT_SIZE/2)*-1) + "px"},
			{Key: "font-size", Val: formatFloat(FONT_SIZE/2) + "px"},
			{Key: "text-anchor", Val: "middle"},
		},
	})
	group := createSVG("g", createOptions{
		attr: []html.Attribute{
			{Key: "data-point-index", Val: strconv.Itoa(index)},
			{Key: "transform", Val: fmt.Sprintf("translate(%f, %f)", x, y)},
		},
	})
	group.AppendChild(rect)
	group.AppendChild(text)
	return group
}

func datasetDot(x, y, radius float64, color, label string, index int) *html.Node {
	dot := createSVG("circle", createOptions{
		style: map[string]string{
			"fill": color,
		},
		attr: []html.Attribute{
			{Key: "data-point-index", Val: strconv.Itoa(index)},
			{Key: "cx", Val: formatFloat(x)},
			{Key: "cy", Val: formatFloat(y)},
			{Key: "r", Val: formatFloat(radius)},
		},
	})
	if label == "" {
		return dot
	}
	setAttribute(dot, "cx", "0")
	setAttribute(dot, "cy", "0")
	text := createSVG("text", createOptions{
		innerHtml: label,
		attr: []html.Attribute{
			{Key: "class", Val: "data-point-value"},
			{Key: "x", Val: "0"},
			{Key: "y", Val: "0"},
			{Key: "dy", Val: formatFloat((FONT_SIZE/2)*-1-radius) + "px"},
			{Key: "font-size", Val: formatFloat(FONT_SIZE) + "px"},
			{Key: "text-anchor", Val: "middle"},
		},
	})
	group := createSVG("g", createOptions{
		attr: []html.Attribute{
			{Key: "data-point-index", Val: strconv.Itoa(index)},
			{Key: "transform", Val: fmt.Sprintf("translate(%f, %f)", x, y)},
		},
	})
	group.AppendChild(dot)
	group.AppendChild(text)
	return group
}

type pathsOptions struct {
	spline, heatLine, regionFill bool
}

type Meta = struct {
	svgDefs  *html.Node
	zeroLine float64
}

func getPaths(xList, yList []float64, color string, o pathsOptions, m Meta) (path, region *html.Node) {
	var s strings.Builder
	if o.spline {
		s.Reset()
		getSplineCurvePointsStr(&s, xList, yList)
	} else {
		for i := range xList {
			if i != 0 {
				s.WriteByte('L')
			}
			fmt.Fprintf(&s, "%f,%f", xList[i], yList[i])
		}
	}
	pointsStr := s.String()
	path = makePath("M"+pointsStr, "line-graph-path", color, "", "")
	if o.heatLine {
		gradientId := makeGradient(m.svgDefs, color, false)
		setStyle(path, "stroke", fmt.Sprintf("url(#%s)", gradientId))
	}
	if o.regionFill {
		gradientId := makeGradient(m.svgDefs, color, true)
		pathStr := fmt.Sprintf("M%f,%fL%sL%f,%f",
			xList[0], m.zeroLine, pointsStr, xList[len(xList)-1], m.zeroLine,
		)
		region = makePath(pathStr, "region-fill", "none", fmt.Sprintf("url(#%s)", gradientId), "")
	}
	return
}

func getSplineCurvePointsStr(o io.Writer, xList, yList []float64) {
	type Point struct {
		x, y float64
	}
	points := make([]*Point, len(xList))
	for i := range xList {
		points[i] = &Point{
			x: xList[i],
			y: yList[i],
		}
	}
	smoothing := 0.2

	line := func(a, b *Point) (length, angle float64) {
		x := b.x - a.x
		y := b.y - a.y
		length = math.Sqrt(math.Pow(x, 2)) + math.Pow(y, 2)
		angle = math.Atan2(y, x)
		return
	}
	controlPoint := func(current, previous, next *Point, reverse bool) *Point {
		p := previous
		if p == nil {
			p = current
		}
		n := next
		if n == nil {
			n = current
		}
		length, angle := line(p, n)
		if reverse {
			angle += math.Pi
		}
		length *= smoothing
		return &Point{
			x: current.x + math.Cos(angle)*length,
			y: current.y + math.Sin(angle)*length,
		}
	}

	bezierCommand := func(o io.Writer, point *Point, i int, a []*Point) {
		cps := controlPoint(a[i-1], a[i-2], point, false)
		cpe := controlPoint(point, a[i-1], a[i+1], true)
		fmt.Fprintf(o, "C %f,%f %f,%f %f,%f",
			cps.x, cps.y, cpe.x, cpe.y, point.x, point.y,
		)
	}

	for i, point := range points {
		if i == 0 {
			fmt.Fprintf(o, "%f,%f", point.x, point.y)
			continue
		}
		o.Write([]byte(" "))
		bezierCommand(o, point, i, points)
	}

}
func setAttribute(n *html.Node, key, value string) {
	for i := range n.Attr {
		if n.Attr[i].Key == key {
			n.Attr[i].Val = value
			return
		}
	}
	n.Attr = append(n.Attr, html.Attribute{Key: key, Val: value})
}

func setStyle(n *html.Node, key, value string) {
	for i := range n.Attr {
		if n.Attr[i].Key == "style" {
			n.Attr[i].Val += key + ":" + value + ";"
			return
		}
	}
	n.Attr = append(n.Attr, html.Attribute{Key: "style", Val: key + ":" + value + ";"})
}

func getAttribute(n *html.Node, key string) string {
	for i := range n.Attr {
		if n.Attr[i].Key == key {
			return n.Attr[i].Val
		}
	}
	return ""
}
