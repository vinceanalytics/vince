// Copyright ©2016 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package graph

import (
	"bytes"
	"os"
	"testing"
)

func TestTrend(t *testing.T) {
	var b bytes.Buffer
	err := Trend(271, 51, []float64{1.0, 10.0, 8.0, 4.0, 5.0}, &b)
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile("trend.svg", b.Bytes(), 0600)
}
