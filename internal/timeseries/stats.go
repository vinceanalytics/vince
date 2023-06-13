package timeseries

import (
	"fmt"
	"net/url"
	"sort"

	"github.com/vinceanalytics/vince/internal/query"
	"github.com/vinceanalytics/vince/pkg/timex"
)

type Stats struct {
	Period timex.Duration
	Domain string
	Key    string
	Prop   Property
	Result query.QueryResult
}

type Plot struct {
	Metric Metric
	Prop   Property
	Values []uint32
	Sum    uint32
	Count  string
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
