package metrics

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	*prometheus.Registry
	Events Events
	Query  Query
}

type Events struct {
	Accepted, Rejected prometheus.Counter
}

type Query struct {
	QueryDuration prometheus.Histogram
	Query         prometheus.Counter
	QueryError    prometheus.Counter
}

func Open(ctx context.Context) context.Context {
	m := &Metrics{
		Registry: prometheus.NewRegistry(),
		Events: Events{
			Accepted: prometheus.NewCounter(prometheus.CounterOpts{
				Namespace: "vince",
				Name:      "events_accepted_total",
				Help:      "Total number of analytics events accepted",
			}),
			Rejected: prometheus.NewCounter(prometheus.CounterOpts{
				Namespace: "vince",
				Name:      "events_rejected_total",
				Help:      "Total number of analytics events rejected",
			}),
		},
		Query: Query{
			QueryDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
				Namespace: "vince",
				Name:      "query_duration",
				Help:      "Query execution time time in seconds",
				Buckets:   prometheus.DefBuckets,
			}),
			Query: prometheus.NewCounter(prometheus.CounterOpts{
				Namespace: "vince",
				Name:      "query_total",
				Help:      "Total number of successful queries",
			}),
			QueryError: prometheus.NewCounter(prometheus.CounterOpts{
				Namespace: "vince",
				Name:      "query_error_total",
				Help:      "Total number of queries that returned error",
			}),
		},
	}
	m.MustRegister(
		m.Events.Accepted, m.Events.Rejected,
		m.Query.Query, m.Query.QueryDuration, m.Query.QueryError,
	)
	return context.WithValue(ctx, metricsKey{}, m)
}

type metricsKey struct{}

type registryKey struct{}

func Get(ctx context.Context) *Metrics {
	return ctx.Value(metricsKey{}).(*Metrics)
}

func New() http.Handler {
	return promhttp.Handler()
}
