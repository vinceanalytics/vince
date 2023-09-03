package api

import (
	"context"
	"log/slog"
	"net/http"

	apiv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/api/v1"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/entry"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/metrics"
	"github.com/vinceanalytics/vince/internal/pj"
	"github.com/vinceanalytics/vince/internal/remoteip"
	"github.com/vinceanalytics/vince/internal/worker"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Events accepts events payloads from vince client script.
func Events(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Set("X-Content-Type-Options", "nosniff")
	req := &entry.Request{}
	err := pj.UnmarshalDefault(req, r.Body)
	if err != nil {
		slog.Error("failed parsing request data", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	req.Ip = remoteip.Get(r)
	req.Ua = r.UserAgent()
	m := metrics.Get(ctx)

	if !accept(ctx, req.D) {
		m.Events.Rejected.Inc()
		w.Header().Set("x-vince-dropped", "1")
		w.WriteHeader(http.StatusOK)
		return
	}
	m.Events.Accepted.Inc()
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
	m := metrics.Get(ctx)
	if !accept(ctx, req.D) {
		m.Events.Rejected.Inc()
		return &emptypb.Empty{}, nil
	}
	m.Events.Accepted.Inc()
	worker.Submit(ctx, req)
	return nil, nil
}
