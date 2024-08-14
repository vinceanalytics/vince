package ro2

import (
	"slices"
	"time"

	"github.com/vinceanalytics/vince/internal/roaring"
	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
)

var bitDepth = map[uint32]uint64{
	timestampField: 64,
	idField:        64,
	bounceField:    1,
	sessionField:   1,
	viewField:      1,
	durationField:  64,
	cityField:      32,
}

func (d *Data) Read(tx *Tx, date, shard uint64,
	match *roaring64.Bitmap, metrics ...string) {
	d.ReadFields(tx, date, shard, match, metricsToProject(metrics)...)
}

func (d *Data) ReadFields(tx *Tx, date, shard uint64,
	match *roaring64.Bitmap, fields ...uint32) {
	for i := range fields {
		f := fields[i]
		b := d.get(f)

		if f == 31 {
			// special handling of events
			it := match.Iterator()
			for it.HasNext() {
				b.SetValue(it.Next(), 1)
			}
			continue
		}
		if f <= cityField {
			tx.ExtractBSI(date, shard, uint64(f), bitDepth[f], match, func(row uint64, c int64) {
				b.SetValue(row, c)
			})
			continue
		}
		// string fields
		tx.ExtractMutex(shard, uint64(f), match, func(row uint64, c *roaring.Container) {
			c.Each(func(u uint16) bool {
				b.SetValue(uint64(u), int64(row))
				return true
			})
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
	case "events":
		return float64(a.Events(foundSet))
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
	b := a.get(idField)
	if foundSet == nil {
		foundSet = b.GetExistenceBitmap()
	}
	return b.IntersectAndTranspose(0, foundSet).GetCardinality()
}

func (a *Data) Visits(foundSet *roaring64.Bitmap) uint64 {
	b := a.get(sessionField)
	if foundSet == nil {
		foundSet = b.GetExistenceBitmap()
	}
	sum, _ := b.Sum(foundSet)
	return uint64(sum)
}

func (a *Data) View(foundSet *roaring64.Bitmap) uint64 {
	b := a.get(viewField)
	if foundSet == nil {
		foundSet = b.GetExistenceBitmap()
	}
	sum, _ := b.Sum(foundSet)
	return uint64(sum)
}

func (a *Data) Duration(foundSet *roaring64.Bitmap) uint64 {
	b := a.get(durationField)
	if foundSet == nil {
		foundSet = b.GetExistenceBitmap()
	}
	sum, _ := b.Sum(foundSet)
	return uint64(sum)
}

func (a *Data) Bounce(foundSet *roaring64.Bitmap) uint64 {
	b := a.get(bounceField)
	if foundSet == nil {
		foundSet = b.GetExistenceBitmap()
	}
	sum, _ := b.Sum(foundSet)
	if sum < 0 {
		return 0
	}
	return uint64(sum)
}

func (a *Data) Events(foundSet *roaring64.Bitmap) uint64 {
	o := a.get(eventsField).GetExistenceBitmap()
	if foundSet != nil {
		o.Add(fieldOffset)
	}
	return o.GetCardinality()
}

func metricsToProject(mets []string) []uint32 {
	m := map[uint32]struct{}{}
	for _, v := range mets {
		switch v {
		case "visitors":
			m[idField] = struct{}{}
		case "visits":
			m[sessionField] = struct{}{}
		case "pageviews":
			m[viewField] = struct{}{}
		case "views_per_visit":
			m[viewField] = struct{}{}
			m[sessionField] = struct{}{}
		case "bounce_rate":
			m[viewField] = struct{}{}
			m[sessionField] = struct{}{}
		case "visit_duration":
			m[durationField] = struct{}{}
			m[sessionField] = struct{}{}
		case "events":
			m[sessionField] = struct{}{}
		}
	}
	o := make([]uint32, 0, len(m))
	for k := range m {
		o = append(o, k)
	}
	slices.Sort(o)
	return o
}
