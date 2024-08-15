package query

import (
	"encoding/json"
	"net/url"

	"github.com/vinceanalytics/vince/internal/ro2"
)

type Query struct {
	period Date
	filter ro2.Filter
}

func New(db *ro2.Store, u url.Values) *Query {
	var fs []Filter
	json.Unmarshal([]byte(u.Get("filters")), &fs)
	period := period(u.Get("period"), u.Get("date"))
	ls := make(ro2.List, len(fs))
	for i := range fs {
		ls[i] = fs[i].To(db)
	}
	return &Query{
		period: period,
		filter: ls,
	}
}

func (q *Query) Start() int64       { return q.period.Start.UnixMilli() }
func (q *Query) End() int64         { return q.period.End.UnixMilli() }
func (q *Query) Filter() ro2.Filter { return q.filter }
func (q *Query) Metrics() []string  { return []string{} }
