package query

import (
	"encoding/json"
	"net/url"

	"github.com/gernest/len64/internal/oracle"
)

type Query struct {
	period Date
	filter oracle.Filter
}

func New(u url.Values) *Query {
	var fs []Filter
	json.Unmarshal([]byte(u.Get("filters")), &fs)
	period := period(u.Get("period"), u.Get("date"))
	ls := make([]oracle.Filter, len(fs))
	for i := range fs {
		ls[i] = fs[i].To()
	}
	return &Query{
		period: period,
		filter: oracle.NewAnd(ls...),
	}
}

func (q *Query) Start() int64          { return q.period.Start.UnixMilli() }
func (q *Query) End() int64            { return q.period.End.UnixMilli() }
func (q *Query) Filter() oracle.Filter { return q.filter }
func (q *Query) Metrics() []string     { return []string{} }
