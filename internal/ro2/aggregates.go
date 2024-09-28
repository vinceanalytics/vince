package ro2

import (
	"math"
	"slices"
	"time"

	"github.com/RoaringBitmap/roaring/v2/roaring64"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
)

type Stats struct {
	Visitors, Visits, PageViews, ViewsPerVisits, BounceRate, VisitDuration float64
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
	tx *Tx,
	shard uint64,
	view uint64,
	match *roaring64.Bitmap, fields ...v1.Field) (err error) {
	for i := range fields {
		f := fields[i]
		switch f {
		case v1.Field_view:
			count, err := tx.Count(shard, view, f, match)
			if err != nil {
				return err
			}
			d.PageViews += float64(count)

		case v1.Field_session:
			count, err := tx.Count(shard, view, f, match)
			if err != nil {
				return err
			}
			d.Visits += float64(count)

		case v1.Field_bounce:
			sum, err := tx.Sum(shard, view, f, match)
			if err != nil {
				return err
			}
			d.BounceRate += float64(sum)
		case v1.Field_duration:
			sum, err := tx.Sum(shard, view, f, match)
			if err != nil {
				return err
			}
			d.VisitDuration += float64(sum)
		case v1.Field_id:
			uniq, err := tx.Unique(shard, view, f, match)
			if err != nil {
				return err
			}
			d.Visitors += float64(uniq)
		}
	}
	return
}

func MetricsToProject(mets []string) []v1.Field {
	m := map[v1.Field]struct{}{}
	for _, v := range mets {
		switch v {
		case "visitors":
			m[v1.Field_id] = struct{}{}
		case "visits":
			m[v1.Field_session] = struct{}{}
		case "pageviews":
			m[v1.Field_view] = struct{}{}
		case "views_per_visit":
			m[v1.Field_view] = struct{}{}
			m[v1.Field_session] = struct{}{}
		case "bounce_rate":
			m[v1.Field_bounce] = struct{}{}
			m[v1.Field_session] = struct{}{}
		case "visit_duration":
			m[v1.Field_duration] = struct{}{}
			m[v1.Field_session] = struct{}{}
		}
	}
	o := make([]v1.Field, 0, len(m))
	for k := range m {
		o = append(o, k)
	}
	slices.Sort(o)
	return o
}
