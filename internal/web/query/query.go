package query

import (
	"encoding/json"
	"net/url"
	"time"

	"github.com/vinceanalytics/vince/internal/ro2"
)

type Query struct {
	period Period
	cmp    *Period
	filter ro2.Filter
	metric string
}

func New(db *ro2.Store, u url.Values) *Query {
	var fs []Filter
	json.Unmarshal([]byte(u.Get("filters")), &fs)
	period := period(u.Get("period"), u.Get("date"))
	if i := u.Get("interval"); i != "" {
		switch i {
		case "minute":
			period.Interval = Minute
		case "hour":
			period.Interval = Hour
		case "date":
			period.Interval = Date
		case "week":
			period.Interval = Week
		case "month":
			period.Interval = Month
		}
	}
	ls := make(ro2.List, len(fs))
	for i := range fs {
		ls[i] = fs[i].To(db)
	}
	var cmp *Period
	switch u.Get("period") {
	case "all", "realtime":
	default:
		now := time.Now().UTC()
		switch u.Get("comparison") {
		case "previous_period":
			diff := period.End.Sub(period.Start)
			cmp = &Period{Start: period.Start.Add(-diff), End: period.End.Add(-diff)}
		case "year_over_year":
			start := period.Start.AddDate(-1, 0, 0)
			end := earliest(period.End, now).AddDate(-1, 0, 0)
			cmp = &Period{Start: start, End: end}
		case "custom":
		}
	}

	return &Query{
		period: period,
		filter: ls,
		cmp:    cmp,
		metric: u.Get("metric"),
	}
}

func earliest(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func (q *Query) Start() int64       { return q.period.Start.UnixMilli() }
func (q *Query) From() string       { return q.period.Start.Format(time.DateOnly) }
func (q *Query) End() int64         { return q.period.End.UnixMilli() }
func (q *Query) To() string         { return q.period.End.Format(time.DateOnly) }
func (q *Query) Interval() Interval { return q.period.Interval }
func (q *Query) Filter() ro2.Filter { return q.filter }
func (q *Query) Metric() string     { return q.metric }
func (q *Query) Compare() *Period   { return q.cmp }
