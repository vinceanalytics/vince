package timeseries

import (
	"sync"
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/gernest/vince/timex"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MetricSaver struct {
	user, session *roaring64.Bitmap
	prop          [12]HourStats
}

var metricSaverPool = &sync.Pool{
	New: func() any {
		return &MetricSaver{
			user:    roaring64.New(),
			session: roaring64.New(),
		}
	},
}

func (m *MetricSaver) Reset() {
	m.user.Clear()
	m.session.Clear()
	for i := range m.prop {
		m.prop[i].Reset()
	}
}

func (m *MetricSaver) Save(
	hr int,
	ts time.Time, f func(u, s *roaring64.Bitmap, h *HourStats),
) {
	h := &m.prop[hr]
	if h.Properties == nil {
		h.Properties = &Properties{}
	}
	f(m.user, m.session, &m.prop[hr])
	ts = timex.Date(ts)
	ts = ts.Add(time.Duration(hr) * time.Hour)
	m.prop[hr].Timestamp = timestamppb.New(ts)
}

func (m *MetricSaver) UpdateHourTotals(hr int, el EntryList) {
	m.prop[hr].Aggregate = el.Aggregate(m.user, m.session)
}

func (m *MetricSaver) Release() {
	m.Reset()
	metricSaverPool.Put(m)
}
