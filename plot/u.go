package plot

import (
	"bytes"
	"encoding/json"
	htpl "html/template"
	"strconv"
	"text/template"
	"time"
)

// U renders javascript snippet plotting graphs using uPlot.js library.
type U struct {
	// This is the site id
	ID            uint64
	Width, Height uint
	Start         time.Time
	Series        []float64
	Class         string
}

var basic = template.Must(
	template.New("").Parse(basicTpl),
)

const basicTpl = `<script defer type="module">
(function () {
	'use strict';
	const x = {{.X}};
	const y = {{.Y}};
	const opts = {{.O}};
	const el = document.getElementById("{{.ID}}")
	opts.width = el.scrollWidth;
	let u = new uPlot(opts, [x, y], el);
})();
</script>`

func (u *U) SparkLine() (htpl.HTML, error) {
	x := make([]int64, len(u.Series))
	for i := range x {
		// for spark lines we don't really care about actual x values. No need to
		// derive unix timestamps.
		x[i] = int64(i)
	}
	o := O{
		Width:   u.Width,
		Height:  u.Height,
		Class:   u.Class,
		PXAlign: false,
		Cursor:  &Cursor{Show: false},
		Select:  &Select{BBox: BBox{Show: false}},
		Legend:  &Legend{Show: false},
		Axes: []Axis{
			{Show: false},
			{Show: false},
		},
		Series: []Series{
			{},
			{
				Stroke: "red",
				Fill:   "rgba(255,0,0,0.1)",
			},
		},
	}
	ctx := map[string]any{
		"X":  mustMarshal(x),
		"Y":  mustMarshal(u.Series),
		"ID": strconv.FormatUint(u.ID, 10),
		"O":  mustMarshal(o),
	}
	var b bytes.Buffer
	err := basic.Execute(&b, ctx)
	if err != nil {
		return "", err
	}
	return htpl.HTML(b.String()), nil

}

func mustMarshal(o interface{}) string {
	b, err := json.Marshal(o)
	if err != nil {
		panic("failed marshalling a plot object " + err.Error())
	}
	return string(b)
}

type O struct {
	Mode    Mode     `json:"mode,omitempty"`
	Title   string   `json:"title,omitempty"`
	ID      string   `json:"id,omitempty"`
	Class   string   `json:"class,omitempty"`
	Width   uint     `json:"width,omitempty"`
	Height  uint     `json:"height,omitempty"`
	Data    TSData   `json:"data,omitempty"`
	PXAlign bool     `json:"pxAlign,omitempty"`
	Series  []Series `json:"series,omitempty"`
	Axes    []Axis   `json:"axes,omitempty"`
	Select  *Select  `json:"select,omitempty"`
	Legend  *Legend  `json:"legend,omitempty"`
	Cursor  *Cursor  `json:"cursor,omitempty"`
}

type Series struct {
	Show    bool   `json:"show,omitempty"`
	Class   string `json:"class,omitempty"`
	Scale   string `json:"scale,omitempty"`
	Auto    bool   `json:"auto,omitempty"`
	PXAlign bool   `json:"pxAlign,omitempty"`
	Width   uint   `json:"width,omitempty"`
	Stroke  string `json:"stroke,omitempty"`
	Fill    string `json:"fill,omitempty"`
}

type Axis struct {
	Show bool `json:"show"`
}

type BBox struct {
	Show bool `json:"show"`
}

type Select struct {
	BBox
	Over bool `json:"over,omitempty"`
}

type Legend struct {
	Show bool `json:"show"`
}

type Cursor struct {
	Show bool `json:"show"`
}

type TSData [][]float64

type Mode uint8

const (
	Aligned Mode = 1 + iota
	Faceted
)
