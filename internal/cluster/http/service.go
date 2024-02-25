package http

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/cluster/rtls"
	"github.com/vinceanalytics/vince/version"
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

	Committed(ctx context.Context) (uint64, error)

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

const (

	// VersionHTTPHeader is the HTTP header key for the version.
	VersionHTTPHeader = "X-RQLITE-VERSION"

	// ServedByHTTPHeader is the HTTP header used to report which
	// node (by node Raft address) actually served the request if
	// it wasn't served by this node.
	ServedByHTTPHeader = "X-RQLITE-SERVED-BY"

	// AllowOriginHeader is the HTTP header for allowing CORS compliant access from certain origins
	AllowOriginHeader = "Access-Control-Allow-Origin"

	// AllowMethodsHeader is the HTTP header for supporting the correct methods
	AllowMethodsHeader = "Access-Control-Allow-Methods"

	// AllowHeadersHeader is the HTTP header for supporting the correct request headers
	AllowHeadersHeader = "Access-Control-Allow-Headers"

	// AllowCredentialsHeader is the HTTP header for supporting specifying credentials
	AllowCredentialsHeader = "Access-Control-Allow-Credentials"
)

type Service struct {
	svr  http.Server
	addr string
	ln   net.Listener

	store   Store
	cluster Cluster

	AllowOrigin string // Value to set for Access-Control-Allow-Origin

	start      time.Time
	lastBackup time.Time

	CACertFile   string // Path to x509 CA certificate used to verify certificates.
	CertFile     string // Path to server's own x509 certificate.
	KeyFile      string // Path to server's own x509 private key.
	ClientVerify bool   // Whether client certificates should verified.
	tls          *tls.Config

	creds CredentialStore

	close chan struct{}

	log *slog.Logger
}

// New returns an uninitialized HTTP service. If credentials is nil, then
// the service performs no authentication and authorization checks.
func New(addr string, store Store, cluster Cluster, credentials CredentialStore) *Service {
	return &Service{
		addr:    addr,
		store:   store,
		cluster: cluster,
		start:   time.Now(),
		creds:   credentials,
		log:     slog.Default().With("component", "http-service"),
	}
}

// Start starts the service.
func (s *Service) Start(ctx context.Context) error {
	s.svr = http.Server{
		Handler:     s,
		BaseContext: func(l net.Listener) context.Context { return ctx },
	}

	var ln net.Listener
	var err error
	if s.CertFile == "" || s.KeyFile == "" {
		ln, err = net.Listen("tcp", s.addr)
		if err != nil {
			return err
		}
	} else {
		mTLSState := rtls.MTLSStateDisabled
		if s.ClientVerify {
			mTLSState = rtls.MTLSStateEnabled
		}
		s.tls, err = rtls.CreateServerConfig(s.CertFile, s.KeyFile, s.CACertFile, mTLSState)
		if err != nil {
			return err
		}
		ln, err = tls.Listen("tcp", s.addr, s.tls)
		if err != nil {
			return err
		}
		var b strings.Builder
		b.WriteString(fmt.Sprintf("secure HTTPS server enabled with cert %s, key %s", s.CertFile, s.KeyFile))
		if s.CACertFile != "" {
			b.WriteString(fmt.Sprintf(", CA cert %s", s.CACertFile))
		}
		if s.ClientVerify {
			b.WriteString(", mutual TLS enabled")
		} else {
			b.WriteString(", mutual TLS disabled")
		}
		// print the message
		s.log.Info(b.String())
	}
	s.ln = ln

	s.close = make(chan struct{})

	go func() {
		err := s.svr.Serve(s.ln)
		if err != nil {
			s.log.Error("Stopped http service", "addr", s.ln.Addr(), "err", err)
		}
	}()
	s.log.Info("Started service", "addr", s.ln.Addr())
	return nil
}

// Close closes the service.
func (s *Service) Close() {
	s.log.Info("closing HTTP service", "addr", s.ln.Addr())
	s.svr.Shutdown(context.Background())
	s.ln.Close()
}

// HTTPS returns whether this service is using HTTPS.
func (s *Service) HTTPS() bool {
	return s.CertFile != "" && s.KeyFile != ""
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.addBuildVersion(w)
	s.addAllowHeaders(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	params, err := NewQueryParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	switch {
	case strings.HasPrefix(r.URL.Path, "/api/v1/stats/realtime/visitors"):
		s.handleRealtime(w, r, params)
	case strings.HasPrefix(r.URL.Path, "/api/v1/stats/aggregate"):
		s.handleAggregate(w, r, params)
	case strings.HasPrefix(r.URL.Path, "/api/v1/stats/timeseries"):
		s.handleTimeseries(w, r, params)
	case strings.HasPrefix(r.URL.Path, "/api/v1/stats/breakdown"):
		s.handleBreakdown(w, r, params)
	case strings.HasPrefix(r.URL.Path, "/api/v1/event"):
		s.handleApiEvent(w, r, params)
	case strings.HasPrefix(r.URL.Path, "/api/event"):
		s.handleEvent(w, r, params)
	case strings.HasPrefix(r.URL.Path, "/backup"):
		s.handleBackup(w, r, params)
	case strings.HasPrefix(r.URL.Path, "/load"):
		s.handleLoad(w, r, params)
	case strings.HasPrefix(r.URL.Path, "/boot"):
		s.handleBoot(w, r, params)
	case strings.HasPrefix(r.URL.Path, "/nodes"):
		s.handleNodes(w, r, params)
	case strings.HasPrefix(r.URL.Path, "/remove"):
		s.handleRemove(w, r, params)
	case strings.HasPrefix(r.URL.Path, "/status"):
		s.handleStatus(w, r, params)
	case strings.HasPrefix(r.URL.Path, "/readyz"):
		s.handleReady(w, r, params)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s *Service) handleRealtime(w http.ResponseWriter, r *http.Request, params QueryParams) {
}
func (s *Service) handleAggregate(w http.ResponseWriter, r *http.Request, params QueryParams) {
}
func (s *Service) handleTimeseries(w http.ResponseWriter, r *http.Request, params QueryParams) {
}
func (s *Service) handleBreakdown(w http.ResponseWriter, r *http.Request, params QueryParams) {
}
func (s *Service) handleApiEvent(w http.ResponseWriter, r *http.Request, params QueryParams) {
}
func (s *Service) handleEvent(w http.ResponseWriter, r *http.Request, params QueryParams) {
}
func (s *Service) handleBackup(w http.ResponseWriter, r *http.Request, params QueryParams) {}
func (s *Service) handleLoad(w http.ResponseWriter, r *http.Request, params QueryParams)   {}
func (s *Service) handleBoot(w http.ResponseWriter, r *http.Request, params QueryParams)   {}
func (s *Service) handleNodes(w http.ResponseWriter, r *http.Request, params QueryParams)  {}
func (s *Service) handleRemove(w http.ResponseWriter, r *http.Request, params QueryParams) {}
func (s *Service) handleStatus(w http.ResponseWriter, r *http.Request, params QueryParams) {}
func (s *Service) handleReady(w http.ResponseWriter, r *http.Request, params QueryParams)  {}

// addBuildVersion adds the build version to the HTTP response.
func (s *Service) addBuildVersion(w http.ResponseWriter) {
	w.Header().Add(VersionHTTPHeader, version.VERSION)
}

// addAllowHeaders adds the Access-Control-Allow-Origin, Access-Control-Allow-Methods,
// and Access-Control-Allow-Headers headers to the HTTP response.
func (s *Service) addAllowHeaders(w http.ResponseWriter) {
	if s.AllowOrigin != "" {
		w.Header().Add(AllowOriginHeader, s.AllowOrigin)
	}
	w.Header().Add(AllowMethodsHeader, "OPTIONS, GET, POST")
	if s.creds == nil {
		w.Header().Add(AllowHeadersHeader, "Content-Type")
	} else {
		w.Header().Add(AllowHeadersHeader, "Content-Type, Authorization")
		w.Header().Add(AllowCredentialsHeader, "true")
	}
}
