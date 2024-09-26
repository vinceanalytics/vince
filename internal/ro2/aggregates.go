package ro2

import (
	"math"
	"slices"
	"time"

	"github.com/gernest/rows"
	"github.com/vinceanalytics/vince/internal/alicia"
	"github.com/vinceanalytics/vince/internal/rbf"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/boolean"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/bsi"
	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
)

func (d *Data) Read(
	rtx *rbf.Tx,
	view string,
	shard uint64,
	match *rows.Row, metrics ...string) {
	d.ReadFields(rtx, view, shard, match, MetricsToProject(metrics)...)
}

func (d *Data) ReadFields(
	rtx *rbf.Tx,
	view string,
	shard uint64,
	match *rows.Row, fields ...alicia.Field) (err error) {
	for i := range fields {
		f := fields[i]
		b := d.mustGet(f)
		switch f {
		case alicia.VIEW:
			err = viewCu(rtx, "view"+view, func(rCu *rbf.Cursor) error {
				return boolean.True(rCu, shard, match, b.SetValue)
			})
		case alicia.SESSION:
			err = viewCu(rtx, "session"+view, func(rCu *rbf.Cursor) error {
				return boolean.True(rCu, shard, match, b.SetValue)
			})
		case alicia.BOUNCE:
			err = viewCu(rtx, "session"+view, func(rCu *rbf.Cursor) error {
				return boolean.Bounce(rCu, shard, match, b.SetValue)
			})
		case alicia.DURATION:
			err = viewCu(rtx, "duration"+view, func(rCu *rbf.Cursor) error {
				return bsi.Extract(rCu, shard, match, b.SetValue)
			})
		case alicia.CITY:
			err = viewCu(rtx, "city"+view, func(rCu *rbf.Cursor) error {
				return bsi.Extract(rCu, shard, match, b.SetValue)
			})
		}
		if err != nil {
			return
		}
	}
	return
}

type Stats struct {
	Visitors, Visits, PageViews, ViewsPerVisits, BounceRate, VisitDuration float64
	uid                                                                    roaring64.Bitmap
}

func (s *Stats) Compute() {
	s.Visitors = float64(s.uid.GetCardinality())
	if s.VisitDuration != 0 {
		s.VisitDuration = time.Duration(s.VisitDuration).Seconds()
	}
	if s.Visits != 0 {
		s.ViewsPerVisits /= s.Visits
		s.ViewsPerVisits = math.Floor(s.ViewsPerVisits)
		s.BounceRate /= s.Visits
		s.BounceRate = math.Floor(s.BounceRate * 100)
		s.VisitDuration /= s.Visits
		s.VisitDuration = math.Floor(s.VisitDuration)
	}

}

func (d *Stats) ReadFields(
	rtx *rbf.Tx,
	view string,
	shard uint64,
	match *rows.Row, fields ...alicia.Field) (err error) {
	for i := range fields {
		f := fields[i]
		switch f {
		case alicia.VIEW:
			err = viewCu(rtx, "view"+view, func(rCu *rbf.Cursor) error {
				return boolean.Count(rCu, shard, match, func(value int64) error {
					d.PageViews += float64(value)
					return nil
				})
			})
		case alicia.SESSION:
			err = viewCu(rtx, "session"+view, func(rCu *rbf.Cursor) error {
				return boolean.Count(rCu, shard, match, func(value int64) error {
					d.Visits += float64(value)
					return nil
				})
			})
		case alicia.BOUNCE:
			err = viewCu(rtx, "bounce"+view, func(rCu *rbf.Cursor) error {
				return boolean.BounceCount(rCu, shard, match, func(value int64) error {
					d.BounceRate += float64(value)
					return nil
				})
			})
		case alicia.DURATION:
			err = viewCu(rtx, "duration"+view, func(rCu *rbf.Cursor) error {
				return bsi.Sum(rCu, match, func(count int32, sum int64) error {
					d.VisitDuration += float64(sum)
					return nil
				})
			})
		case alicia.ID:
			err = viewCu(rtx, "id"+view, func(rCu *rbf.Cursor) error {
				return bsi.Extract(rCu, shard, match, func(column uint64, value int64) {
					d.uid.Add(uint64(value))
				})
			})
		}
		if err != nil {
			return
		}
	}
	return
}

func (a *Data) Stats(foundSet *roaring64.Bitmap) (o Stats) {
	o.Visitors = float64(a.Visitors(foundSet))
	o.Visits = float64(a.Visits(foundSet))
	o.PageViews = float64(a.View(foundSet))
	if o.Visits != 0 {
		o.ViewsPerVisits = o.PageViews / o.Visits
		o.ViewsPerVisits = math.Floor(o.ViewsPerVisits)
	}
	o.BounceRate = float64(a.Bounce(foundSet))
	if o.Visits != 0 {
		o.BounceRate /= o.Visits
		o.BounceRate = math.Floor(o.BounceRate * 100)
	}
	o.VisitDuration = time.Duration(a.Duration(foundSet)).Seconds()
	if o.Visits != 0 {
		o.VisitDuration /= o.Visits
		o.VisitDuration = math.Floor(o.VisitDuration)
	}
	return
}

func (a *Data) Compute(metric string, foundSet *roaring64.Bitmap) float64 {
	switch metric {
	case "visitors":
		return float64(a.Visitors(foundSet))
	case "visits":
		return float64(a.Visits(foundSet))
	case "pageviews":
		return float64(a.View(foundSet))
	case "views_per_visit":
		views := float64(a.View(foundSet))
		visits := float64(a.Visits(foundSet))
		r := float64(0)
		if visits != 0 {
			r = views / visits
		}
		return r
	case "bounce_rate":
		bounce := float64(a.Bounce(foundSet))
		visits := float64(a.Visits(foundSet))
		r := float64(0)
		if visits != 0 {
			r = bounce / visits
		}
		return r
	case "visit_duration":
		duration := a.Duration(foundSet)
		visits := float64(a.Visits(foundSet))
		r := float64(0)
		if visits != 0 && duration != 0 {
			d := time.Duration(duration) * time.Millisecond
			r = d.Seconds() / visits
		}
		return r
	default:
		return 0
	}
}

func (a *Data) Visitors(foundSet *roaring64.Bitmap) uint64 {
	b := a.get(alicia.ID)
	if b == nil {
		return 0
	}
	if foundSet == nil {
		foundSet = b.GetExistenceBitmap()
	}
	return b.IntersectAndTranspose(0, foundSet).GetCardinality()
}

func (a *Data) Visits(foundSet *roaring64.Bitmap) uint64 {
	b := a.get(alicia.SESSION)
	if b == nil {
		return 0
	}
	if foundSet == nil {
		foundSet = b.GetExistenceBitmap()
	}
	sum, _ := b.Sum(foundSet)
	return uint64(sum)
}

func (a *Data) View(foundSet *roaring64.Bitmap) uint64 {
	b := a.get(alicia.VIEW)
	if b == nil {
		return 0
	}
	if foundSet == nil {
		foundSet = b.GetExistenceBitmap()
	}
	sum, _ := b.Sum(foundSet)
	return uint64(sum)
}

func (a *Data) Duration(foundSet *roaring64.Bitmap) uint64 {
	b := a.get(alicia.DURATION)
	if b == nil {
		return 0
	}
	if foundSet == nil {
		foundSet = b.GetExistenceBitmap()
	}
	sum, _ := b.Sum(foundSet)
	return uint64(sum)
}

func (a *Data) Bounce(foundSet *roaring64.Bitmap) uint64 {
	b := a.get(alicia.BOUNCE)
	if b == nil {
		return 0
	}
	if foundSet == nil {
		foundSet = b.GetExistenceBitmap()
	}
	sum, _ := b.Sum(foundSet)
	if sum < 0 {
		return 0
	}
	return uint64(sum)
}

func MetricsToProject(mets []string) []alicia.Field {
	m := map[alicia.Field]struct{}{}
	for _, v := range mets {
		switch v {
		case "visitors":
			m[alicia.ID] = struct{}{}
		case "visits":
			m[alicia.SESSION] = struct{}{}
		case "pageviews":
			m[alicia.VIEW] = struct{}{}
		case "views_per_visit":
			m[alicia.VIEW] = struct{}{}
			m[alicia.SESSION] = struct{}{}
		case "bounce_rate":
			m[alicia.BOUNCE] = struct{}{}
			m[alicia.SESSION] = struct{}{}
		case "visit_duration":
			m[alicia.DURATION] = struct{}{}
			m[alicia.SESSION] = struct{}{}
		case "events":
			m[alicia.SESSION] = struct{}{}
		}
	}
	o := make([]alicia.Field, 0, len(m))
	for k := range m {
		o = append(o, k)
	}
	slices.Sort(o)
	return o
}
