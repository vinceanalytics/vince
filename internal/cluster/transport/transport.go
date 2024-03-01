// Package transport provides a Transport for github.com/hashicorp/raft over gRPC.
package transport

import (
	"sync/atomic"
	"time"

	"github.com/hashicorp/raft"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/cluster/connections"
	"google.golang.org/grpc"
)

type Manager struct {
	*connections.Manager
	rpcChan          chan raft.RPC
	heartbeatFunc    atomic.Value
	heartbeatTimeout time.Duration
}

// New creates both components of raft-grpc-transport: a gRPC service and a Raft Transport.
func New(conns *connections.Manager, options ...Option) *Manager {
	m := &Manager{
		Manager: conns,
		rpcChan: make(chan raft.RPC),
	}
	for _, opt := range options {
		opt(m)
	}
	return m
}

// Register the RaftTransport gRPC service on a gRPC server.
func (m *Manager) Register(s grpc.ServiceRegistrar) {
	v1.RegisterRaftTransportServer(s, gRPCAPI{manager: m})
}

// Transport returns a raft.Transport that communicates over gRPC.
func (m *Manager) Transport() raft.Transport {
	return &raftAPI{manager: m}
}

func (m *Manager) Close() (err error) {
	m.Manager.Close()
	return
}
