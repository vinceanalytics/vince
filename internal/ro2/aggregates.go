package ro2

import (
	"slices"
	"time"

	"github.com/vinceanalytics/vince/internal/alicia"
	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
)

func (d *Data) Read(tx *Tx, shard uint64,
	match *roaring64.Bitmap, metrics ...string) {
	d.ReadFields(tx, shard, match, MetricsToProject(metrics)...)
}

func (d *Data) ReadFields(tx *Tx, shard uint64,
	match *roaring64.Bitmap, fields ...alicia.Field) {
	for i := range fields {
		f := fields[i]
		b := d.get(f)
		tx.ExtractBSI(shard, uint64(f), match, func(row uint64, c int64) {
			b.SetValue(row, c)
		})
	}
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
	if foundSet == nil {
		foundSet = b.GetExistenceBitmap()
	}
	return b.IntersectAndTranspose(0, foundSet).GetCardinality()
}

func (a *Data) Visits(foundSet *roaring64.Bitmap) uint64 {
	b := a.get(alicia.SESSION)
	if foundSet == nil {
		foundSet = b.GetExistenceBitmap()
	}
	sum, _ := b.Sum(foundSet)
	return uint64(sum)
}

func (a *Data) View(foundSet *roaring64.Bitmap) uint64 {
	b := a.get(alicia.VIEW)
	if foundSet == nil {
		foundSet = b.GetExistenceBitmap()
	}
	sum, _ := b.Sum(foundSet)
	return uint64(sum)
}

func (a *Data) Duration(foundSet *roaring64.Bitmap) uint64 {
	b := a.get(alicia.DURATION)
	if foundSet == nil {
		foundSet = b.GetExistenceBitmap()
	}
	sum, _ := b.Sum(foundSet)
	return uint64(sum)
}

func (a *Data) Bounce(foundSet *roaring64.Bitmap) uint64 {
	b := a.get(alicia.BOUNCE)
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
