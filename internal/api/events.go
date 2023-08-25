package api

import (
	"context"
	"net/http"

	"log/slog"

	"github.com/dgraph-io/badger/v4"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/entry"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/remoteip"
	"github.com/vinceanalytics/vince/internal/worker"
)

// Events accepts events payloads from vince client script.
func Events(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	xlg := slog.Default().WithGroup("api").
		With("path", r.URL.Path, "method", r.Method)

	req := entry.NewRequest()
	defer req.Release()

	err := req.Parse(r.Body)
	if err != nil {
		xlg.Error("failed parsing request data", "err", err)
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
	worker.Submit(ctx, req)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// returns true if domain is events are accepted. CHecks is done to only ensure
// there is an existing site registered with the domain.
func accept(ctx context.Context, domain string) bool {
	return db.Get(ctx).View(func(txn *badger.Txn) error {
		key := keys.Site(domain)
		_, err := txn.Get([]byte(key))
		return err
	}) == nil
}
