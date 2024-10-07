package ro2

import (
	"math"
	"time"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/fieldset"
	"github.com/vinceanalytics/vince/internal/roaring"
)

type Stats struct {
	Visitors, Visits, PageViews, ViewsPerVisits, BounceRate, VisitDuration float64
	uid                                                                    *roaring.Bitmap
}

func NewStats(fs fieldset.Set) *Stats {
	var s Stats
	if fs.Has(v1.Field_id) {
		s.uid = roaring.NewBitmap()
	}
	return &s
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
	if s.uid != nil {
		s.Visitors = float64(s.uid.GetCardinality())
	}
	if s.VisitDuration != 0 {
		s.VisitDuration = time.Duration(s.VisitDuration).Seconds()
	}
	s.ViewsPerVisits = s.PageViews
	if s.Visits != 0 {
		s.ViewsPerVisits /= s.Visits
		s.ViewsPerVisits = math.Round(s.ViewsPerVisits)
		s.BounceRate /= s.Visits
		s.BounceRate = math.Floor(s.BounceRate * 100)
		s.VisitDuration /= s.Visits
		s.VisitDuration = math.Floor(s.VisitDuration)
	}
}

func (d *Stats) Read(tx *Tx, shard, view uint64, match *roaring.Bitmap, fields fieldset.Set) error {
	return fields.Each(func(f v1.Field) (err error) {
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
			uniq, err := tx.Transpose(shard, view, f, match)
			if err != nil {
				return err
			}
			d.uid.Or(uniq)
		}
		return
	})
}
