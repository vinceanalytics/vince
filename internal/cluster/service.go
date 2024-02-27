package cluster

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
)

const (
	// MuxRaftHeader is the byte used to indicate internode Raft communications.
	MuxRaftHeader = 1

	// MuxClusterHeader is the byte used to request internode cluster state information.
	MuxClusterHeader = 2 // Cluster state communications
)

// Dialer is the interface dialers must implement.
type Dialer interface {
	// Dial is used to create a connection to a service listening
	// on an address.
	Dial(address string, timeout time.Duration) (net.Conn, error)
}

type Database interface {
	Data(ctx context.Context, req *v1.Data) error
	Realtime(ctx context.Context, req *v1.Realtime_Request) (*v1.Realtime_Response, error)
	Aggregate(ctx context.Context, req *v1.Aggregate_Request) (*v1.Aggregate_Response, error)
	Timeseries(ctx context.Context, req *v1.Timeseries_Request) (*v1.Timeseries_Response, error)
	Breakdown(ctx context.Context, req *v1.BreakDown_Request) (*v1.BreakDown_Response, error)
	Load(ctx context.Context, req *v1.Load_Request) error
}

// Manager is the interface node-management systems must implement
type Manager interface {
	// LeaderAddr returns the Raft address of the leader of the cluster.
	LeaderAddr() (string, error)

	// CommitIndex returns the Raft commit index of the cluster.
	CommitIndex() (uint64, error)

	// Remove removes the node, given by id, from the cluster
	Remove(context.Context, *v1.RemoveNode_Request) error

	// Join joins a remote node to the cluster.
	Join(context.Context, *v1.Join_Request) error
}

// CredentialStore is the interface credential stores must support.
type CredentialStore interface {
	// AA authenticates and checks authorization for the given perm.
	AA(username, password string, perm v1.Credential_Permission) bool
}

// Service provides information about the node and cluster.
type Service struct {
	ln   net.Listener // Incoming connections to the service
	addr net.Addr     // Address on which this service is listening

	db  Database // The queryable system.
	mgr Manager  // The cluster management system.

	credentialStore CredentialStore

	mu      sync.RWMutex
	https   bool   // Serving HTTPS?
	apiAddr string // host:port this node serves the HTTP API.

	logger *slog.Logger
}

// New returns a new instance of the cluster service
func New(ln net.Listener, db Database, m Manager, credentialStore CredentialStore) *Service {
	return &Service{
		ln:              ln,
		addr:            ln.Addr(),
		db:              db,
		mgr:             m,
		logger:          slog.Default().With("component", "cluster"),
		credentialStore: credentialStore,
	}
}

// Close closes the service.
func (s *Service) Close() error {
	s.ln.Close()
	return nil
}

// Addr returns the address the service is listening on.
func (s *Service) Addr() string {
	return s.addr.String()
}

// EnableHTTPS tells the cluster service the API serves HTTPS.
func (s *Service) EnableHTTPS(b bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.https = b
}

// SetAPIAddr sets the API address the cluster service returns.
func (s *Service) SetAPIAddr(addr string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.apiAddr = addr
}

// GetAPIAddr returns the previously-set API address
func (s *Service) GetAPIAddr() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.apiAddr
}

// GetNodeAPIURL returns fully-specified HTTP(S) API URL for the
// node running this service.
func (s *Service) GetNodeAPIURL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	scheme := "http"
	if s.https {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, s.apiAddr)
}

func (s *Service) checkCommandPerm(c *v1.Credentials, perm v1.Credential_Permission) bool {
	if s.credentialStore == nil {
		return true
	}
	return s.credentialStore.AA(c.GetUsername(), c.GetPassword(), perm)
}
