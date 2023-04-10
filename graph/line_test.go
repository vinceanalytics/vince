// Copyright ©2016 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package graph

import (
	"os"
	"testing"
)

func TestTrend(t *testing.T) {
	s := SiteTrend()
	os.WriteFile("trend.svg", []byte(s), 0600)
}
