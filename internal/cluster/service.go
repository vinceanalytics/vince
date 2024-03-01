package cluster

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/buffers"
	"github.com/vinceanalytics/vince/internal/cluster/store"
	"google.golang.org/protobuf/types/known/emptypb"
)

// CredentialStore is the interface credential stores must support.
type CredentialStore interface {
	// AA authenticates and checks authorization for the given perm.
	AA(username, password string, perm v1.Credential_Permission) bool
}

// Service provides information about the node and cluster.
type Service struct {
	v1.UnsafeInternalCLusterServer
	addr    net.Addr // Address on which this service is listening
	store   store.Storage
	mu      sync.RWMutex
	https   bool   // Serving HTTPS?
	apiAddr string // host:port this node serves the HTTP API.
	logger  *slog.Logger
}

var _ v1.InternalCLusterServer = (*Service)(nil)

// New returns a new instance of the cluster service
func New(db store.Storage) *Service {
	return &Service{
		store:  db,
		logger: slog.Default().With("component", "cluster"),
	}
}

func (s *Service) Join(ctx context.Context, req *v1.Join_Request) (*v1.Join_Response, error) {
	err := s.store.Join(ctx, req)
	if err != nil {
		return nil, err
	}
	leader, err := s.store.LeaderAddr(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.Join_Response{Leader: leader}, nil
}

func (s *Service) Load(ctx context.Context, req *v1.Load_Request) (*emptypb.Empty, error) {
	err := s.store.Load(ctx, req)
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
	err := s.store.Backup(svr.Context(), req, cs)
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
	err := s.store.Remove(ctx, req)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) Notify(ctx context.Context, req *v1.Notify_Request) (*emptypb.Empty, error) {
	err := s.store.Notify(ctx, req)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) NodeAPI(ctx context.Context, req *v1.NodeAPIRequest) (*v1.NodeMeta, error) {
	idx, err := s.store.CommitIndex(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.NodeMeta{
		Url:         s.GetNodeAPIURL(),
		CommitIndex: idx,
	}, nil
}

func (s *Service) SendData(ctx context.Context, req *v1.Data) (*emptypb.Empty, error) {
	err := s.store.Data(ctx, req)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
func (s *Service) Realtime(ctx context.Context, req *v1.Realtime_Request) (*v1.Realtime_Response, error) {
	return s.store.Realtime(ctx, req)
}

func (s *Service) Aggregate(ctx context.Context, req *v1.Aggregate_Request) (*v1.Aggregate_Response, error) {
	return s.store.Aggregate(ctx, req)
}
func (s *Service) Timeseries(ctx context.Context, req *v1.Timeseries_Request) (*v1.Timeseries_Response, error) {
	return s.store.Timeseries(ctx, req)
}
func (s *Service) BreakDown(ctx context.Context, req *v1.BreakDown_Request) (*v1.BreakDown_Response, error) {
	return s.store.Breakdown(ctx, req)
}

// Close closes the service.
func (s *Service) Close() error {
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
