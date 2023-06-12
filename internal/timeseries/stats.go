package timeseries

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/vinceanalytics/vince/pkg/property"
	"github.com/vinceanalytics/vince/pkg/timex"
)

type Stats struct {
	Start, End time.Time
	Period     timex.Duration
	Domain     string
	Key        string
	Prop       Property
	Timestamps []int64
	Aggregate  Aggregate
	Timeseries PropertiesResult
}

type Plot struct {
	Metric Metric
	Prop   Property
	Values []uint32
	Sum    uint32
	Count  string
}

func (s *Stats) Series() []*Plot {
	o := s.Timeseries[s.Prop.String()]
	var r [VisitDurations + 1]*Plot
	for k := range r {
		r[k] = &Plot{
			Metric: property.Metric(k),
			Prop:   s.Prop,
		}
		p := r[k]
		x, ok := o[p.Metric.String()]
		if !ok {
			continue
		}
		p.Values = x[s.Key]
		p.Sum = sum(p.Values)
	}
	for _, v := range r {
		switch v.Metric {
		case BounceRates:
			if v.Sum != 0 && r[Visits].Sum > 0 {
				f := float64(v.Sum) / float64(r[Visits].Sum)
				f *= 100
				v.Count = strconv.FormatFloat(
					f, 'f', 1, 64,
				) + "%"
			} else {
				v.Count = "0"
			}
		case VisitDurations:
			if v.Sum != 0 && r[Visits].Sum > 0 {
				f := float64(v.Sum) / float64(r[Visits].Sum)
				d := time.Duration(f)
				v.Count = d.String()
			} else {
				v.Count = "0"
			}
		default:
			v.Count = strconv.Itoa(int(v.Sum))
		}
	}
	return r[:]
}

func (s *Stats) ActiveSeries() []uint32 {
	return s.Timeseries[s.Prop.String()][property.Visitors.String()][s.Key]
}

func (s *Stats) Value(o OutValue) []uint32 {
	return o[s.Key]
}

func (s *Stats) QueryPeriod(period timex.Duration) string {
	q := make(url.Values)
	q.Set("w", period.String())
	q.Set("k", s.Key)
	q.Set("p", s.Prop.String())
	return fmt.Sprintf("/%s/stats?%s", url.PathEscape(s.Domain), q.Encode())
}

func (s *Stats) QueryProp(prop, metric, key string) string {
	q := make(url.Values)
	q.Set("w", s.Period.String())
	q.Set("k", key)
	q.Set("p", prop)
	return fmt.Sprintf("/%s/stats?%s", url.PathEscape(s.Domain), q.Encode())
}

func (s *Stats) PlotTime() (string, error) {
	b, err := json.Marshal(s.Timestamps)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (s *Stats) Count(metric string) uint32 {
	o := s.Aggregate[s.Prop.String()][metric]
	for _, v := range o {
		if v.Key == s.Key {
			return v.Value
		}
	}
	return 0
}

type Panel struct {
	Stats   *Stats
	Prop    string
	Metrics AggregateMetricsStatValue
}

func (s *Stats) Panel(prop string) Panel {
	return Panel{
		Stats:   s,
		Prop:    prop,
		Metrics: s.Aggregate[prop],
	}
}

func (s *Stats) PlotValue(metric string) (string, error) {
	o := s.Timeseries[s.Prop.String()][metric][s.Key]
	if len(o) == 0 {
		o = make([]uint32, len(s.Timeseries))
	}
	b, err := json.Marshal(o)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

type Aggregate map[string]AggregateMetricsStatValue

// AggregateValue maps keys to value for a specific metric
type AggregateValue map[string]uint32

type AggregateMetricsStatValue map[string]StatList

type StatValue struct {
	Key   string
	Value uint32
}

func (s StatValue) Icon() string {
	source := s.Key
	if source == "" {
		source = "placeholder"
	}
	return "/favicon/sources/" + url.PathEscape(s.Key)
}

var _ sort.Interface = (*StatList)(nil)

type StatList []StatValue

func (s StatList) Len() int {
	return len(s)
}
func (s StatList) Less(i, j int) bool {
	return s[i].Value < s[j].Value
}

func (s StatList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func sum(ls []uint32) (o uint32) {
	for _, v := range ls {
		o += v
	}
	return
}

type RootOptions struct {
	Metric Metric        `json:"metric,omitempty"`
	Prop   Property      `json:"prop,omitempty"`
	Key    string        `json:"key,omitempty"`
	Start  time.Time     `json:"start,omitempty"`
	Window time.Duration `json:"window,omitempty"`
	Offset time.Duration `json:"offset,omitempty"`
	// When set to true and Prop is Base, other props will not be queried. This is
	// useful to query only base aggregates . Like the one used on sites
	// index.
	NoProps bool `json:"noProps,omitempty"`
}

func Root(ctx context.Context, uid, sid uint64, opts RootOptions) (o Stats) {
	if opts.Prop == Base {
		opts.Key = BaseKey
	}
	if opts.Window == 0 {
		opts.Window = time.Hour * 24
	}
	q := Query(ctx, QueryRequest{
		UserID: uid,
		SiteID: sid,
		BaseQuery: BaseQuery{
			Window:  opts.Window,
			Offset:  opts.Offset,
			Start:   opts.Start,
			Metrics: allMetrics(opts),
			Filters: allProperties(opts),
		},
	})

	o.Start = q.Start
	o.End = q.End

	o.Timestamps = q.Timestamps

	o.Aggregate = make(Aggregate)
	o.Timeseries = q.Result
	for k, v := range q.Result {
		am := make(AggregateMetricsStatValue)
		for m, mv := range v {
			ls := make(StatList, 0, len(mv))
			for ok, ov := range mv {
				st := StatValue{
					Key:   ok,
					Value: sum(ov),
				}
				ls = append(ls, st)
			}
			// sort in ascending order.
			sort.Sort(sort.Reverse(ls))
			am[m] = ls
		}
		o.Aggregate[k] = am
	}
	return
}

var allMetricsLs = []Metric{
	Visitors,
	Views,
	Events,
	Visits,
	BounceRates,
	VisitDurations,
}

func allMetrics(opt RootOptions) []Metric {
	if opt.Prop == Base {
		if opt.NoProps {
			return []Metric{
				opt.Metric,
			}
		}
		return allMetricsLs
	}
	return []Metric{
		opt.Metric,
	}
}

func allProperties(opt RootOptions) FilterList {
	if opt.Prop != Base || opt.NoProps {
		// No need to select other properties if its not for the base. This is the
		// case when we are searching based on a key
		return []*Filter{
			{Property: opt.Prop, Expr: FilterExpr{
				Text: opt.Key,
			}},
		}
	}
	o := make([]*Filter, City+1)
	a := make([]Metric, 0, VisitDurations+1)
	for i := Visitors; i <= VisitDurations; i++ {
		if i == opt.Metric {
			continue
		}
		a = append(a, i)
	}

	for i := range o {
		if i == int(Base) {
			o[i] = &Filter{
				Property: Property(i),
				Expr: FilterExpr{
					Text: BaseKey,
				},
			}
		} else {
			o[i] = &Filter{
				Property: Property(i),
				Omit:     a,
				Expr: FilterExpr{
					Text: "*",
				},
			}
		}
	}
	return o
}
