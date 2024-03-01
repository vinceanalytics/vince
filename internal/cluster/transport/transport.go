// Package transport provides a Transport for github.com/hashicorp/raft over gRPC.
package transport

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/hashicorp/raft"
	"github.com/pkg/errors"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"google.golang.org/grpc"
)

var (
	errCloseErr = errors.New("error closing connections")
)

type Manager struct {
	localAddress raft.ServerAddress
	dialOptions  []grpc.DialOption

	rpcChan          chan raft.RPC
	heartbeatFunc    atomic.Value
	heartbeatTimeout time.Duration
	connections      sync.Map
}

// New creates both components of raft-grpc-transport: a gRPC service and a Raft Transport.
func New(localAddress raft.ServerAddress, dialOptions []grpc.DialOption, options ...Option) *Manager {
	m := &Manager{
		localAddress: localAddress,
		dialOptions:  dialOptions,
		rpcChan:      make(chan raft.RPC),
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
	return raftAPI{m}
}

func (m *Manager) Close() (err error) {
	m.connections.Range(func(key, value any) bool {
		e := value.(*conn).Close()
		if e != nil {
			err = e
		}
		return true
	})
	return
}
