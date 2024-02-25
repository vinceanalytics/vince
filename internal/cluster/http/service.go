package http

import (
	"context"
	"errors"
	"io"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
)

var (
	// ErrLeaderNotFound is returned when a node cannot locate a leader
	ErrLeaderNotFound = errors.New("leader not found")
)

type ResultsError interface {
	Error() string
	IsAuthorized() bool
}

type Database interface {
	KV(ctx context.Context, req *v1.KV_Request) error
	Realtime(ctx context.Context, req *v1.Realtime_Request) (*v1.Realtime_Response, error)
	Aggregate(ctx context.Context, req *v1.Aggregate_Request) (*v1.Aggregate_Response, error)
	Timeseries(ctx context.Context, req *v1.Timeseries_Request) (*v1.Timeseries_Response, error)
	Breakdown(ctx context.Context, req *v1.BreakDown_Request) (*v1.BreakDown_Response, error)
	Load(ctx context.Context, req *v1.Load_Request) error
}

// Store is the interface the Raft-based database must implement.
type Store interface {
	Database

	// Remove removes the node from the cluster.
	Remove(ctx context.Context, rn *v1.RemoveNode_Request) error

	// LeaderAddr returns the Raft address of the leader of the cluster.
	LeaderAddr(ctx context.Context) (string, error)

	// Ready returns whether the Store is ready to service requests.
	Ready(ctx context.Context) bool

	// Nodes returns the slice of store.Servers in the cluster
	Nodes(ctx context.Context) (*v1.Server_List, error)

	// Backup writes backup of the node state to dst
	Backup(ctx context.Context, br *v1.Backup_Request, dst io.Writer) error

	// ReadFrom reads and loads a SQLite database into the node, initially bypassing
	// the Raft system. It then triggers a Raft snapshot, which will then make
	// Raft aware of the new data.
	ReadFrom(ctx context.Context, r io.Reader) (int64, error)
}

// GetAddresser is the interface that wraps the GetNodeAPIAddr method.
// GetNodeAPIAddr returns the HTTP API URL for the node at the given Raft address.
type GetAddresser interface {
	GetNodeAPIAddr(ctx context.Context, addr string) (string, error)
}

type Cluster interface {
	GetAddresser

	KV(ctx context.Context, nodeAddr string, cred *v1.Credential, req *v1.KV_Request) error
	Load(ctx context.Context, nodeAddr string, cred *v1.Credential, req *v1.Load_Request) error
	Remove(ctx context.Context, nodeAddr string, cred *v1.Credential, req *v1.RemoveNode_Request) error
	Backup(ctx context.Context, br *v1.Backup_Request, nodeAddr string, cred *v1.Credential, dst io.Writer) error

	Realtime(ctx context.Context, nodeAddr string, cred *v1.Credential, req *v1.Realtime_Request) (*v1.Realtime_Response, error)
	Aggregate(ctx context.Context, nodeAddr string, cred *v1.Credential, req *v1.Aggregate_Request) (*v1.Aggregate_Response, error)
	Timeseries(ctx context.Context, nodeAddr string, cred *v1.Credential, req *v1.Timeseries_Request) (*v1.Timeseries_Response, error)
	Breakdown(ctx context.Context, nodeAddr string, cred *v1.Credential, req *v1.BreakDown_Request) (*v1.BreakDown_Response, error)
}

// CredentialStore is the interface credential stores must support.
type CredentialStore interface {
	// AA authenticates and checks authorization for the given perm.
	AA(username, password string, perm v1.Credential_Permission) bool
}
