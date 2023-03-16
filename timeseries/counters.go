package timeseries

import (
	"sync"
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/gernest/vince/timex"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MetricSaver struct {
	bloom *roaring64.Bitmap
	prop  [12]HourStats
}

var metricSaverPool = &sync.Pool{
	New: func() any {
		return new(MetricSaver)
	},
}

func (m *MetricSaver) Reset() {
	m.bloom.Clear()
	for i := range m.prop {
		m.prop[i].Reset()
	}
}

func (m *MetricSaver) Save(
	hr int,
	ts time.Time, f func(*roaring64.Bitmap, *HourStats),
) {
	h := &m.prop[hr]
	if h.Properties == nil {
		h.Properties = &Properties{}
	}
	f(m.bloom, &m.prop[hr])
	ts = timex.Date(ts)
	ts = ts.Add(time.Duration(hr) * time.Hour)
	m.prop[hr].Timestamp = timestamppb.New(ts)
}

func (m *MetricSaver) UpdateHourTotals(hr int, el EntryList) {
	m.prop[hr].Visitors = el.Visitors(m.bloom)
	m.prop[hr].Visits = el.Visits(m.bloom)
}

func (m *MetricSaver) Release() {
	m.Reset()
	metricSaverPool.Put(m)
}
