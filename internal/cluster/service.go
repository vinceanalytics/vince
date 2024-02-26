package cluster

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"time"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	pb "google.golang.org/protobuf/proto"
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

// Open opens the Service.
func (s *Service) Open() error {
	go s.serve()
	s.logger.Info("service listening", "addr", s.addr)
	return nil
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

func (s *Service) serve() error {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return err
		}

		go s.handleConn(conn)
	}
}

func (s *Service) checkCommandPerm(c *v1.Command_Request, perm v1.Credential_Permission) bool {
	if s.credentialStore == nil {
		return true
	}

	username := ""
	password := ""
	if c.Credentials != nil {
		username = c.Credentials.GetUsername()
		password = c.Credentials.GetPassword()
	}
	return s.credentialStore.AA(username, password, perm)
}

const unauthorized = "unauthorized"

func (s *Service) handleConn(conn net.Conn) {
	defer conn.Close()
	ctx := context.Background()
	b := make([]byte, protoBufferLengthSize)
	for {
		_, err := io.ReadFull(conn, b)
		if err != nil {
			return
		}
		sz := binary.LittleEndian.Uint64(b[0:])

		p := make([]byte, sz)
		_, err = io.ReadFull(conn, p)
		if err != nil {
			return
		}

		c := &v1.Command_Request{}
		err = pb.Unmarshal(p, c)
		if err != nil {
			conn.Close()
		}
		marshalAndWrite(conn, s.handle(ctx, c))
	}
}

func (s *Service) handle(ctx context.Context, req *v1.Command_Request) *v1.Command_Response {
	switch e := req.Request.(type) {
	case *v1.Command_Request_Join:
		if !s.checkCommandPerm(req, v1.Credential_JOIN) {
			return &v1.Command_Response{Response: &v1.Command_Response_Join{
				Join: &v1.Join_Response{Error: unauthorized},
			}}
		}
		_ = e
	case *v1.Command_Request_NodeApi:
		if !s.checkCommandPerm(req, v1.Credential_NODE_API) {
			return &v1.Command_Response{Response: &v1.Command_Response_NodeApi{
				NodeApi: &v1.NodeAPI_Response{Error: unauthorized},
			}}
		}
	case *v1.Command_Request_Load:
		if !s.checkCommandPerm(req, v1.Credential_LOAD) {
			return &v1.Command_Response{Response: &v1.Command_Response_Load{
				Load: &v1.Load_Response{Error: unauthorized},
			}}
		}
	case *v1.Command_Request_Backup:
		if !s.checkCommandPerm(req, v1.Credential_BACKUP) {
			return &v1.Command_Response{Response: &v1.Command_Response_Backup{
				Backup: &v1.Backup_Response{Error: unauthorized},
			}}
		}
	case *v1.Command_Request_Data:
		if !s.checkCommandPerm(req, v1.Credential_DATA) {
			return &v1.Command_Response{Response: &v1.Command_Response_Data{
				Data: &v1.DataService_Response{Error: unauthorized},
			}}
		}
	case *v1.Command_Request_Query:
		if !s.checkCommandPerm(req, v1.Credential_QUERY) {
			return &v1.Command_Response{Response: &v1.Command_Response_Query{
				Query: &v1.Query_Response{Error: unauthorized},
			}}
		}
	}
	return nil
}

func marshalAndWrite(conn net.Conn, m pb.Message) {
	p, err := pb.Marshal(m)
	if err != nil {
		conn.Close()
	}
	writeBytesWithLength(conn, p)
}

func writeBytesWithLength(conn net.Conn, p []byte) {
	b := make([]byte, protoBufferLengthSize)
	binary.LittleEndian.PutUint64(b[0:], uint64(len(p)))
	conn.Write(b)
	conn.Write(p)
}
