package api

import (
	"context"
	"net/http"

	"github.com/dgraph-io/badger/v4"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/entry"
	"github.com/vinceanalytics/vince/internal/events"
	"github.com/vinceanalytics/vince/internal/log"
	"github.com/vinceanalytics/vince/internal/remoteip"
	"github.com/vinceanalytics/vince/internal/timeseries"
	v1 "github.com/vinceanalytics/vince/proto/v1"
)

// Events accepts events payloads from vince client script.
func Events(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("X-Content-Type-Options", "nosniff")
	xlg := log.Get()
	req := entry.NewRequest()
	defer req.Release()

	err := req.Parse(r.Body)
	if err != nil {
		xlg.Err(err).Msg("Failed decoding json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	req.IP = remoteip.Get(r)
	req.UserAgent = r.UserAgent()
	ctx := r.Context()
	if !accept(ctx, req.Domain) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	e, err := events.Parse(req, core.Now(r.Context()))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	timeseries.Register(ctx, e)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// returns true if domain is events are accepted. CHecks is done to only ensure
// there is an existing site registered with the domain.
func accept(ctx context.Context, domain string) bool {
	return db.Get(ctx).View(func(txn *badger.Txn) error {
		key := (&v1.StoreKey{
			Prefix: v1.StorePrefix_SITES,
			Key:    domain,
		}).Badger()
		_, err := txn.Get([]byte(key))
		return err
	}) == nil
}
