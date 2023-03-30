// Copyright ©2016 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package graph

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/wcharczuk/go-chart"
)

func TestOther(t *testing.T) {
	now := time.Now()
	graph := chart.Chart{
		Width:  1613,
		Height: 240,
		Series: []chart.Series{
			chart.TimeSeries{
				Style: chart.Style{
					Show:        true,
					StrokeColor: TrendStroke, // will supercede defaults
					FillColor:   TrendFill,   // will supercede defaults
				},
				XValues: []time.Time{
					now.Add(time.Hour),
					now.Add(2 * time.Hour),
					now.Add(3 * time.Hour),
					now.Add(4 * time.Hour),
					now.Add(5 * time.Hour),
				},
				YValues: []float64{1.0, 10.0, 8.0, 4.0, 5.0},
			},
		},
	}
	buffer := bytes.NewBuffer([]byte{})
	err := graph.Render(chart.PNG, buffer)
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile("new.png", buffer.Bytes(), 0600)
}
