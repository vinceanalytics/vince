package oracle

import (
	"slices"
	"time"

	"github.com/RoaringBitmap/roaring/v2/roaring64"
	"github.com/gernest/len64/internal/rbf"
	"github.com/gernest/len64/internal/rbf/cursor"
	"github.com/gernest/rows"
	"go.etcd.io/bbolt"
)

type Aggregate struct {
	Results map[string]Value `json:"results"`
}

type Value struct {
	Value float64 `json:"value"`
}

func (o *Oracle) Aggregate(start, end int64, domain string, filter Filter, metrics []string) (*Aggregate, error) {
	m := newAggregate()
	err := o.db.Select(start, end, domain, filter, func(rTx *rbf.Tx, tx *bbolt.Tx, shard uint64, match *rows.Row) error {
		return m.read(rTx, shard, match, metrics...)
	})
	if err != nil {
		return nil, err
	}
	a := &Aggregate{
		Results: make(map[string]Value),
	}
	for _, k := range metrics {
		a.Results[k] = Value{Value: m.Compute(k)}
	}
	return a, nil
}

type aggregate struct {
	id       *roaring64.Bitmap
	uid      *roaring64.BSI
	session  *roaring64.BSI
	view     *roaring64.BSI
	bounce   *roaring64.BSI
	duration *roaring64.BSI
}

func newAggregate() *aggregate {
	return &aggregate{
		id:       roaring64.New(),
		uid:      roaring64.NewDefaultBSI(),
		session:  roaring64.NewDefaultBSI(),
		view:     roaring64.NewDefaultBSI(),
		bounce:   roaring64.NewDefaultBSI(),
		duration: roaring64.NewDefaultBSI(),
	}
}

func (a *aggregate) Compute(metric string) float64 {
	switch metric {
	case "visitors":
		return float64(a.Visitors())
	case "visits":
		return float64(a.Visits())
	case "pageviews":
		return float64(a.View())
	case "events":
		return float64(a.Events())
	case "views_per_visit":
		views := float64(a.View())
		visits := float64(a.Visits())
		r := float64(0)
		if visits != 0 {
			r = views / visits
		}
		return r
	case "bounce_rate":
		bounce := float64(a.Bounce())
		visits := float64(a.Visits())
		r := float64(0)
		if visits != 0 {
			r = bounce / visits
		}
		return r
	case "visit_duration":
		duration := a.Duration()
		visits := float64(a.Visits())
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
func (a *aggregate) Visitors() uint64 {
	return a.uid.Transpose().GetCardinality()
}

func (a *aggregate) Visits() uint64 {
	sum, _ := a.session.Sum(a.session.GetExistenceBitmap())
	return uint64(sum)
}

func (a *aggregate) View() uint64 {
	sum, _ := a.view.Sum(a.view.GetExistenceBitmap())
	return uint64(sum)
}

func (a *aggregate) Duration() uint64 {
	sum, _ := a.duration.Sum(a.duration.GetExistenceBitmap())
	return uint64(sum)
}

func (a *aggregate) Bounce() uint64 {
	sum, _ := a.bounce.Sum(a.bounce.GetExistenceBitmap())
	if sum < 0 {
		return 0
	}
	return uint64(sum)
}

func (a *aggregate) Events() uint64 {
	return a.id.GetCardinality()
}

func (a *aggregate) read(rTx *rbf.Tx, shard uint64, match *rows.Row, metrics ...string) error {
	for _, m := range metricsToProject(metrics) {
		switch m {
		case "id":
			a.id.AddMany(match.Columns())
		default:
			err := cursor.Tx(rTx, m, func(c *rbf.Cursor) error {
				return extractBSI(c, shard, match, func(column uint64, value int64) error {
					a.uid.SetValue(column, value)
					return nil
				})
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
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
