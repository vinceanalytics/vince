package api

import (
	"context"
	"net/http"

	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	apiv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/api/v1"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/entry"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/pj"
	"github.com/vinceanalytics/vince/internal/remoteip"
	"github.com/vinceanalytics/vince/internal/worker"
	"google.golang.org/protobuf/types/known/emptypb"
)

var AcceptedEvents = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "vince",
	Name:      "events_accepted_total",
	Help:      "Total number of analytics events accepted",
})

var RejectedEvents = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "vince",
	Name:      "events_rejected_total",
	Help:      "Total number of analytics events rejected",
})

func init() {
	prometheus.MustRegister(AcceptedEvents, RejectedEvents)
}

// Events accepts events payloads from vince client script.
func Events(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	xlg := slog.Default().WithGroup("api").
		With("path", r.URL.Path, "method", r.Method)

	req := &entry.Request{}

	err := pj.UnmarshalDefault(req, r.Body)
	if err != nil {
		xlg.Error("failed parsing request data", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	req.Ip = remoteip.Get(r)
	req.Ua = r.UserAgent()
	ctx := r.Context()
	if !accept(ctx, req.D) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	worker.Submit(ctx, req)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// returns true if domain is events are accepted. CHecks is done to only ensure
// there is an existing site registered with the domain.
func accept(ctx context.Context, domain string) (ok bool) {
	txn := db.Get(ctx).NewTransaction(false)
	key := keys.Site(domain)
	defer key.Release()
	ok = txn.Has(key.Bytes())
	key.Release()
	txn.Close()
	return
}

// SendEvent accepts analytics event. Assumes req has already been validated
func (a *API) SendEvent(ctx context.Context, req *apiv1.Event) (*emptypb.Empty, error) {
	if !accept(ctx, req.D) {
		RejectedEvents.Inc()
		return &emptypb.Empty{}, nil
	}
	AcceptedEvents.Inc()
	worker.Submit(ctx, req)
	return nil, nil
}
