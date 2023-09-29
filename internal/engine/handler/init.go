package handler

import (
	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"
	"go.opentelemetry.io/otel/trace"
)

func NewServer(cfg server.Config, e *sqle.Engine, sb server.SessionBuilder, listener server.ServerEventListener) *Handler {
	var tracer trace.Tracer
	if cfg.Tracer != nil {
		tracer = cfg.Tracer
	} else {
		tracer = sql.NoopTracer
	}

	sm := server.NewSessionManager(sb, tracer, e.Analyzer.Catalog.Database, e.MemoryManager, e.ProcessList, cfg.Address)
	return &Handler{
		e:                 e,
		sm:                sm,
		readTimeout:       cfg.ConnReadTimeout,
		disableMultiStmts: cfg.DisableClientMultiStatements,
		maxLoggedQueryLen: cfg.MaxLoggedQueryLen,
		encodeLoggedQuery: cfg.EncodeLoggedQuery,
		sel:               listener,
	}
}
