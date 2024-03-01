package cluster

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/buffers"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Database interface {
	Data(ctx context.Context, req *v1.Data) error
	Realtime(ctx context.Context, req *v1.Realtime_Request) (*v1.Realtime_Response, error)
	Aggregate(ctx context.Context, req *v1.Aggregate_Request) (*v1.Aggregate_Response, error)
	Timeseries(ctx context.Context, req *v1.Timeseries_Request) (*v1.Timeseries_Response, error)
	Breakdown(ctx context.Context, req *v1.BreakDown_Request) (*v1.BreakDown_Response, error)
	Load(ctx context.Context, req *v1.Load_Request) error
	Backup(ctx context.Context, br *v1.Backup_Request, dst io.Writer) error
}

// Manager is the interface node-management systems must implement
type Manager interface {
	// LeaderAddr returns the Raft address of the leader of the cluster.
	LeaderAddr(ctx context.Context) (string, error)

	// CommitIndex returns the Raft commit index of the cluster.
	CommitIndex(ctx context.Context) (uint64, error)

	// Remove removes the node, given by id, from the cluster
	Remove(context.Context, *v1.RemoveNode_Request) error

	// Join joins a remote node to the cluster.
	Join(context.Context, *v1.Join_Request) error

	Notify(ctx context.Context, nr *v1.Notify_Request) error
}

// CredentialStore is the interface credential stores must support.
type CredentialStore interface {
	// AA authenticates and checks authorization for the given perm.
	AA(username, password string, perm v1.Credential_Permission) bool
}

// Service provides information about the node and cluster.
type Service struct {
	v1.UnsafeInternalCLusterServer
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

var _ v1.InternalCLusterServer = (*Service)(nil)

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

func (s *Service) Join(ctx context.Context, req *v1.Join_Request) (*v1.Join_Response, error) {
	err := s.mgr.Join(ctx, req)
	if err != nil {
		return nil, err
	}
	leader, err := s.mgr.LeaderAddr(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.Join_Response{Leader: leader}, nil
}

func (s *Service) Load(ctx context.Context, req *v1.Load_Request) (*emptypb.Empty, error) {
	err := s.db.Load(ctx, req)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

const chunkSize = 4 << 10

func (s *Service) Backup(req *v1.Backup_Request, svr v1.InternalCLuster_BackupServer) error {
	cs := &chunkedWriter{
		b:   buffers.Bytes(),
		svr: svr,
	}
	defer func() {
		cs.b.Release()
		cs.b = nil
	}()
	err := s.db.Backup(svr.Context(), req, cs)
	if err != nil {
		return err
	}
	if cs.b.Len() != 0 {
		return svr.Send(&v1.Backup_Response{
			Data: cs.b.Bytes(),
		})
	}
	return nil
}

type chunkedWriter struct {
	b   *buffers.BytesBuffer
	svr v1.InternalCLuster_BackupServer
}

func (c *chunkedWriter) Write(p []byte) (int, error) {
	n, err := c.b.Write(p)
	if err != nil {
		return 0, err
	}
	if c.b.Len() >= chunkSize {
		err := c.svr.Send(&v1.Backup_Response{
			Data: c.b.Bytes(),
		})
		if err != nil {
			return 0, err
		}
		c.b.Reset()
	}
	return n, nil
}

func (s *Service) RemoveNode(ctx context.Context, req *v1.RemoveNode_Request) (*emptypb.Empty, error) {
	err := s.mgr.Remove(ctx, req)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) Notify(ctx context.Context, req *v1.Notify_Request) (*emptypb.Empty, error) {
	err := s.mgr.Notify(ctx, req)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) NodeAPI(ctx context.Context, req *v1.NodeAPIRequest) (*v1.NodeMeta, error) {
	idx, err := s.mgr.CommitIndex(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.NodeMeta{
		Url:         s.GetNodeAPIURL(),
		CommitIndex: idx,
	}, nil
}

func (s *Service) SendData(ctx context.Context, req *v1.Data) (*emptypb.Empty, error) {
	err := s.db.Data(ctx, req)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
func (s *Service) Realtime(ctx context.Context, req *v1.Realtime_Request) (*v1.Realtime_Response, error) {
	return s.db.Realtime(ctx, req)
}

func (s *Service) Aggregate(ctx context.Context, req *v1.Aggregate_Request) (*v1.Aggregate_Response, error) {
	return s.db.Aggregate(ctx, req)
}
func (s *Service) Timeseries(ctx context.Context, req *v1.Timeseries_Request) (*v1.Timeseries_Response, error) {
	return s.db.Timeseries(ctx, req)
}
func (s *Service) BreakDown(ctx context.Context, req *v1.BreakDown_Request) (*v1.BreakDown_Response, error) {
	return s.db.Breakdown(ctx, req)
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
