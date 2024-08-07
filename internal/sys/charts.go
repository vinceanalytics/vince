package sys

import "github.com/RoaringBitmap/roaring/v2/roaring64"

type chartBSI struct {
	ts         *roaring64.Bitmap
	ram        *roaring64.BSI
	histograms [3]*roaring64.BSI
	requests   *roaring64.BSI
}

func newChartBSI() *chartBSI {
	o := &chartBSI{
		ts:       roaring64.New(),
		ram:      roaring64.NewDefaultBSI(),
		requests: roaring64.NewDefaultBSI(),
	}
	o.histograms[0] = roaring64.NewDefaultBSI()
	o.histograms[1] = roaring64.NewDefaultBSI()
	o.histograms[2] = roaring64.NewDefaultBSI()
	return o
}
