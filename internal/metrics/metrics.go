package metrics

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	Events Events
}

type Events struct {
	Accepted, Rejected prometheus.Counter
}

func Open(ctx context.Context, reg *prometheus.Registry) context.Context {
	m := &Metrics{
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
	}
	reg.MustRegister(m.Events.Accepted, m.Events.Rejected)
	return context.WithValue(ctx, metricsKey{}, m)
}

type metricsKey struct{}

func Get(ctx context.Context) *Metrics {
	return ctx.Value(metricsKey{}).(*Metrics)
}

func New() http.Handler {
	return promhttp.Handler()
}
