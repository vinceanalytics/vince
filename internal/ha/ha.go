package ha

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/hashicorp/raft"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/must"
)

type haKey struct{}

type Ha struct {
	Raft     *raft.Raft
	Transit  *Transit
	provider *provider
	log      *slog.Logger
}

func (h *Ha) Close() error {
	h.log.Info("closing raft transport")
	return h.Transit.Close()
}

func Get(ctx context.Context) *Ha {
	return ctx.Value(haKey{}).(*Ha)
}

func Open(ctx context.Context) (context.Context, *Ha) {
	base := db.Get(ctx)
	o := config.Get(ctx)
	rc := raft.DefaultConfig()
	rc.LocalID = raft.ServerID(o.ServerId)
	rdb := &DB{db: db.GetRaft(ctx)}
	tr := NewTransport(ctx, raft.ServerAddress(o.ListenAddress), 3*time.Minute)
	ss := must.Must(raft.NewFileSnapshotStore(
		o.RaftPath,
		2, os.Stderr,
	))("failed initializing a file snapshot store")
	r := must.Must(raft.NewRaft(rc, NewFSM(base), rdb, rdb, ss, tr))(
		"failed initializing raft",
	)
	h := &Ha{
		Raft:    r,
		Transit: tr,
		provider: &provider{
			raft: r,
			base: base,
		},
		log: slog.Default().With("component", "ha"),
	}
	ctx = db.Set(ctx, h.provider)
	return context.WithValue(ctx, haKey{}, h), h
}
