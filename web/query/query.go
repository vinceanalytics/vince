package query

import (
	"net/url"

	"github.com/gernest/len64/internal/oracle"
)

type Query struct{}

func New(u url.Values) *Query {
	return &Query{}
}

func (q *Query) Start() int64          { return 0 }
func (q *Query) End() int64            { return 0 }
func (q *Query) Filter() oracle.Filter { return oracle.Noop() }
func (q *Query) Metrics() []string     { return []string{} }
