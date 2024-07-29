package len64

import (
	"slices"
	"time"

	"github.com/RoaringBitmap/roaring/v2/roaring64"
)

type Aggregate struct {
	Results map[string]Value `json:"results"`
}

type Value struct {
	Value float64 `json:"value"`
}

func (db *Store) Aggregate(domain string, start, end time.Time, filter Filter, metrics []string) (*Aggregate, error) {
	match, err := db.Select(start, end, domain, filter, metricsToProject(metrics))
	if err != nil {
		return nil, err
	}
	a := &Aggregate{
		Results: make(map[string]Value),
	}
	for _, k := range metrics {
		a.Results[k] = Value{Value: match.Compute(k)}
	}
	return a, nil
}

func metricsToProject(mets []string) []string {
	m := map[string]struct{}{}
	for _, v := range mets {
		switch v {
		case "visitors":
			m["uid"] = struct{}{}
		case "visits":
			m["session"] = struct{}{}
		case "pageviews":
			m["view"] = struct{}{}
		case "views_per_visit":
			m["view"] = struct{}{}
			m["session"] = struct{}{}
		case "bounce_rate":
			m["view"] = struct{}{}
			m["session"] = struct{}{}
		case "visit_duration":
			m["duration"] = struct{}{}
			m["session"] = struct{}{}
		case "events":
			m["session"] = struct{}{}
		}
	}
	o := make([]string, 0, len(m))
	for k := range m {
		o = append(o, k)
	}
	slices.Sort(o)
	return o
}

type Projection map[string]*roaring64.BSI

func (p Projection) Compute(metric string) float64 {
	switch metric {
	case "visitors":
		return float64(p.Visitors())
	case "visits":
		return float64(p.Visits())
	case "pageviews":
		return float64(p.Views())
	case "events":
		return float64(p.Events())
	case "views_per_visit":
		views := float64(p.Views())
		visits := float64(p.Visits())
		r := float64(0)
		if visits != 0 {
			r = views / visits
		}
		return r
	case "bounce_rate":
		bounce := float64(p.Bounce())
		visits := float64(p.Visits())
		r := float64(0)
		if visits != 0 {
			r = bounce / visits
		}
		return r
	case "visit_duration":
		duration := p.Duration()
		visits := float64(p.Visits())
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
func (p Projection) Visitors() uint64 {
	uniq := p["uid"].Transpose()
	return uniq.GetCardinality()
}

func (p Projection) Events() uint64 {
	return p["session"].GetCardinality()
}

func (p Projection) Views() uint64 {
	b := p["view"]
	sum, _ := b.Sum(b.GetExistenceBitmap())
	return uint64(sum)
}

func (p Projection) Duration() uint64 {
	b := p["duration"]
	sum, _ := b.Sum(b.GetExistenceBitmap())
	return uint64(sum)
}

func (p Projection) Visits() uint64 {
	b := p["session"]
	sum, _ := b.Sum(b.GetExistenceBitmap())
	return uint64(sum)
}

func (p Projection) Bounce() uint64 {
	b := p["bounce"]
	sum, _ := b.Sum(b.GetExistenceBitmap())
	if sum < 0 {
		return 0
	}
	return uint64(sum)
}

type Group struct {
	Key        string
	Value      int64
	Projection Projection
}

func (p Projection) GroupBy(name string) []Group {
	b := p[name]
	uniq := b.Transpose()
	it := uniq.Iterator()
	o := make([]Group, 0, uniq.GetCardinality())
	for it.HasNext() {
		value := int64(it.Next())
		r := b.CompareValue(parallel(), roaring64.EQ, value, value, b.GetExistenceBitmap())
		o = append(o, Group{
			Key:        name,
			Value:      value,
			Projection: p.clone(r, name),
		})
	}
	return o
}

func (p Projection) clone(foundSet *roaring64.Bitmap, skip string) Projection {
	o := make(Projection)
	for k, v := range p {
		if _, ok := p[skip]; ok {
			continue
		}
		b := v.NewBSIRetainSet(foundSet)
		o[k] = b
	}
	return o
}
