package timeseries

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"sort"
	"strconv"
	"time"

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

func (s *Stats) QueryPeriod(period timex.Duration) string {
	q := make(url.Values)
	q.Set("o", period.String())
	q.Set("k", s.Key)
	q.Set("p", s.Prop.String())
	return fmt.Sprintf("/%s/stats?%s", url.PathEscape(s.Domain), q.Encode())
}

type Aggregate map[string]AggregateMetricsStatValue

// AggregateValue maps keys to value for a specific metric
type AggregateValue map[string]FloatValue

type AggregateMetricsStatValue map[string]StatList

type StatValue struct {
	Key   string
	Value FloatValue
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

func sum(ls []float64) (o float64) {
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
		opts.Offset = time.Hour * 24
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
					Value: FloatValue(sum(ov)),
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

type FloatValue float64

func (v FloatValue) String() string {
	f := float64(v)
	p := math.Floor(math.Log10(math.Abs(f)))
	if p <= 2 {
		return strconv.FormatFloat(f, 'f', -1, 64)
	}
	x := math.Floor(p / 3)
	s := math.Pow(10, p-x*3) * +(f / math.Pow(10, p))
	s = math.Round(s*100) / 100
	k := []string{"", "K", "M", "B", "T"}
	return strconv.FormatFloat(s, 'f', -1, 64) + k[int(x)]
}
