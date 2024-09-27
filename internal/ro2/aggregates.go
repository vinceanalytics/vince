package ro2

import (
	"math"
	"slices"
	"time"

	"github.com/RoaringBitmap/roaring/v2/roaring64"
	"github.com/gernest/rows"
	"github.com/vinceanalytics/vince/internal/alicia"
	"github.com/vinceanalytics/vince/internal/rbf"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/boolean"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/bsi"
)

type Stats struct {
	Visitors, Visits, PageViews, ViewsPerVisits, BounceRate, VisitDuration float64
	uid                                                                    roaring64.Bitmap
}

func StatToValue(metric string) func(s *Stats) float64 {
	switch metric {
	case "visitors":
		return func(s *Stats) float64 { return s.Visitors }
	case "visits":
		return func(s *Stats) float64 { return s.Visits }
	case "pageviews":
		return func(s *Stats) float64 { return s.PageViews }
	case "bounce_rate":
		return func(s *Stats) float64 { return s.BounceRate }
	case "views_per_visit":
		return func(s *Stats) float64 { return s.ViewsPerVisits }
	case "visit_duration":
		return func(s *Stats) float64 { return s.VisitDuration }
	default:
		return func(s *Stats) float64 { return 0 }
	}
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
	s.BounceRate = min(s.BounceRate, 100)
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
				return bsi.Extract(rCu, shard, match, func(column uint64, value int64) error {
					d.uid.Add(uint64(value))
					return nil
				})
			})
		}
		if err != nil {
			return
		}
	}
	return
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
