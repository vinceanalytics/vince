package timeseries

import (
	"context"
	"math"
	"strconv"
	"time"
)

type Stats struct {
	Timestamps []int64
	Metrics    map[string]FloatValue
	Props      map[string][]StatValue
}

type StatValue struct {
	Key   string
	Value FloatValue
}

func sum(ls []float64) (o float64) {
	for _, v := range ls {
		o += v
	}
	return
}

func RootQuery(ctx context.Context,
	uid, sid uint64, selectedMetric Metric,
	selectedProp Property,
	key string,
	offset time.Duration,
) (o Stats) {
	q := Query(ctx, QueryRequest{
		UserID: uid,
		SiteID: sid,
		BaseQuery: BaseQuery{
			Offset:  offset,
			Metrics: allMetrics(),
			Filters: allProperties(selectedMetric, selectedProp, key),
		},
	})
	o.Timestamps = q.Timestamps
	o.Metrics = make(map[string]FloatValue)
	// calculate base stats
	base := q.Result[selectedProp.String()]
	for i := Visitors; i <= VisitDurations; i++ {
		o.Metrics[i.String()] = FloatValue(sum(base[i.String()][BaseKey]))
	}

	for i := Event; i <= City; i++ {
		for k, v := range q.Result[i.String()][selectedMetric.String()] {
			o.Props[i.String()] = append(o.Props[i.String()], StatValue{
				Key:   k,
				Value: FloatValue(sum(v)),
			})
		}
	}

	// calculate bounce rate
	visits := o.Metrics[Visits.String()]
	if visits != 0 {
		// avoid dividing by zero, thats why whe check visits != 0
		o.Metrics[BounceRates.String()] =
			(o.Metrics[BounceRates.String()] / visits) * 100
	}
	return
}

func allMetrics() []Metric {
	o := make([]Metric, VisitDurations+1)
	for i := range o {
		o[i] = Metric(i)
	}
	return o
}

func allProperties(selected Metric, selectedProp Property, key string) []*Filter {
	if selectedProp != Base {
		// No need to select other properties if its not for the base. This is the
		// case when we are searching based on a key
		return []*Filter{
			{Property: selectedProp, Expr: FilterExpr{
				Text: key,
			}},
		}
	}
	o := make([]*Filter, City+1)
	a := make([]Metric, 0, VisitDurations+1)
	for i := Visitors; i <= VisitDurations; i++ {
		if i == selected {
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
