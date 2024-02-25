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
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/cluster/rtls"
	"github.com/vinceanalytics/vince/internal/cluster/store"
	"github.com/vinceanalytics/vince/version"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
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

	Status() (*v1.Status_Store, error)
}

// GetAddresser is the interface that wraps the GetNodeAPIAddr method.
// GetNodeAPIAddr returns the HTTP API URL for the node at the given Raft address.
type GetAddresser interface {
	GetNodeAPIAddr(ctx context.Context, addr string) (string, error)
}

type Cluster interface {
	GetAddresser

	KV(ctx context.Context, nodeAddr string, cred *v1.Credential, req *v1.KV_Request) error
	Load(ctx context.Context, nodeAddr string, cred *v1.Credential, req *v1.Load_Request, timeout time.Duration, retries int) error
	Remove(ctx context.Context, nodeAddr string, cred *v1.Credential, timeout time.Duration, req *v1.RemoveNode_Request) error
	Backup(ctx context.Context, br *v1.Backup_Request, nodeAddr string, cred *v1.Credential, timeout time.Duration, dst io.Writer) error

	Realtime(ctx context.Context, nodeAddr string, cred *v1.Credential, req *v1.Realtime_Request) (*v1.Realtime_Response, error)
	Aggregate(ctx context.Context, nodeAddr string, cred *v1.Credential, req *v1.Aggregate_Request) (*v1.Aggregate_Response, error)
	Timeseries(ctx context.Context, nodeAddr string, cred *v1.Credential, req *v1.Timeseries_Request) (*v1.Timeseries_Response, error)
	Breakdown(ctx context.Context, nodeAddr string, cred *v1.Credential, req *v1.BreakDown_Request) (*v1.BreakDown_Response, error)

	Status() (*v1.Status_Cluster, error)
}

// CredentialStore is the interface credential stores must support.
type CredentialStore interface {
	// AA authenticates and checks authorization for the given perm.
	AA(username, password string, perm v1.Credential_Permission) bool
}

const (
	defaultTimeout = 30 * time.Second

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
	case r.URL.Path == "/" || r.URL.Path == "":
		http.Redirect(w, r, "/status", http.StatusFound)
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
func (s *Service) handleNodes(w http.ResponseWriter, r *http.Request, params QueryParams) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if !s.CheckRequestPerm(r, v1.Credential_STATUS) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	servers, err := s.store.Nodes(ctx)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == store.ErrNotOpen {
			statusCode = http.StatusServiceUnavailable
		}
		http.Error(w, fmt.Sprintf("store nodes: %s", err.Error()), statusCode)
		return
	}
	nodes := NewNodesFromServers(servers)
	if !params.NonVoters() {
		nodes = Voters(nodes)
	}
	// Now test the nodes
	lAddr, err := s.store.LeaderAddr(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("leader address: %s", err.Error()),
			http.StatusInternalServerError)
		return
	}
	var wg sync.WaitGroup
	timeout := params.Timeout(defaultTimeout)
	for _, n := range nodes.Items {
		n := n
		wg.Add(1)
		go func() {
			defer wg.Done()
			TestNode(ctx, n, s.cluster, lAddr, timeout)
		}()
	}
	wg.Wait()
	s.write(w, nodes)
}
func (s *Service) handleRemove(w http.ResponseWriter, r *http.Request, params QueryParams) {
	if !s.CheckRequestPerm(r, v1.Credential_REMOVE) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if r.Method != "DELETE" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	rn := &v1.RemoveNode_Request{}
	err = protojson.Unmarshal(b, rn)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	err = s.store.Remove(ctx, rn)
	if err == nil {
		return
	}
	if !errors.Is(err, store.ErrNotLeader) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if s.DoRedirect(w, r, params) {
		return
	}

	addr, err := s.store.LeaderAddr(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("leader address: %s", err.Error()),
			http.StatusInternalServerError)
		return
	}
	if addr == "" {
		http.Error(w, ErrLeaderNotFound.Error(), http.StatusServiceUnavailable)
		return
	}

	username, password, ok := r.BasicAuth()
	if !ok {
		username = ""
	}

	w.Header().Add(ServedByHTTPHeader, addr)
	removeErr := s.cluster.Remove(ctx, addr, makeCredentials(username, password),
		params.Timeout(defaultTimeout), rn)
	if err != nil {
		if removeErr.Error() == "unauthorized" {
			http.Error(w, "remote backup not authorized", http.StatusUnauthorized)
		} else {
			http.Error(w, removeErr.Error(), http.StatusInternalServerError)
		}
		return

	}
}
func (s *Service) handleBackup(w http.ResponseWriter, r *http.Request, params QueryParams) {
	if !s.CheckRequestPerm(r, v1.Credential_BACKUP) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	br := &v1.Backup_Request{
		Leader:   !params.NoLeader(),
		Compress: params.Compress(),
	}
	ctx := r.Context()
	err := s.store.Backup(ctx, br, w)
	if err == nil {
		s.lastBackup = time.Now()
		return
	}
	if !errors.Is(err, store.ErrNotLeader) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if s.DoRedirect(w, r, params) {
		return
	}

	addr, err := s.store.LeaderAddr(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("leader address: %s", err.Error()),
			http.StatusInternalServerError)
		return
	}
	if addr == "" {
		http.Error(w, ErrLeaderNotFound.Error(), http.StatusServiceUnavailable)
		return
	}

	username, password, ok := r.BasicAuth()
	if !ok {
		username = ""
	}

	w.Header().Add(ServedByHTTPHeader, addr)
	backupErr := s.cluster.Backup(ctx, br, addr, makeCredentials(username, password),
		params.Timeout(defaultTimeout), w)
	if err != nil {
		if backupErr.Error() == "unauthorized" {
			http.Error(w, "remote backup not authorized", http.StatusUnauthorized)
		} else {
			http.Error(w, backupErr.Error(), http.StatusInternalServerError)
		}
		return

	}
}
func (s *Service) handleLoad(w http.ResponseWriter, r *http.Request, params QueryParams) {
	if !s.CheckRequestPerm(r, v1.Credential_LOAD) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	lr := &v1.Load_Request{Data: b}
	err = s.store.Load(ctx, lr)
	if err == nil {
		return
	}
	if !errors.Is(err, store.ErrNotLeader) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if s.DoRedirect(w, r, params) {
		return
	}

	addr, err := s.store.LeaderAddr(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("leader address: %s", err.Error()),
			http.StatusInternalServerError)
		return
	}
	if addr == "" {
		http.Error(w, ErrLeaderNotFound.Error(), http.StatusServiceUnavailable)
		return
	}

	username, password, ok := r.BasicAuth()
	if !ok {
		username = ""
	}

	w.Header().Add(ServedByHTTPHeader, addr)
	loadErr := s.cluster.Load(ctx, addr, makeCredentials(username, password), lr,
		params.Timeout(defaultTimeout), params.Retries(0))
	if loadErr != nil {
		if loadErr.Error() == "unauthorized" {
			http.Error(w, "remote load not authorized", http.StatusUnauthorized)
		} else {
			http.Error(w, loadErr.Error(), http.StatusInternalServerError)
		}
		return
	}
}

func (s *Service) doCtx(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, defaultTimeout)
}

func (s *Service) handleBoot(w http.ResponseWriter, r *http.Request, params QueryParams) {
	if !s.CheckRequestPerm(r, v1.Credential_LOAD) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	s.log.Info("Starting booting process")
	_, err := s.store.ReadFrom(r.Context(), r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
}
func (s *Service) handleStatus(w http.ResponseWriter, r *http.Request, params QueryParams) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if !s.CheckRequestPerm(r, v1.Credential_STATUS) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	storeStatus, err := s.store.Status()
	if err != nil {
		http.Error(w, fmt.Sprintf("store stats: %s", err.Error()),
			http.StatusInternalServerError)
		return
	}
	clusterStatus, err := s.cluster.Status()
	if err != nil {
		http.Error(w, fmt.Sprintf("store stats: %s", err.Error()),
			http.StatusInternalServerError)
		return
	}
	status := s.status()
	status.Http.Cluster = clusterStatus
	status.Store = storeStatus
	s.write(w, status)
}

func (s *Service) write(w http.ResponseWriter, msg proto.Message) {
	data, _ := protojson.Marshal(msg)
	_, err := w.Write(data)
	if err != nil {
		s.log.Error("failed writing response data", "err", err)
	}

}

// DoRedirect checks if the request is a redirect, and if so, performs the redirect.
// Returns true caller can consider the request handled. Returns false if the request
// was not a redirect and the caller should continue processing the request.
func (s *Service) DoRedirect(w http.ResponseWriter, r *http.Request, qp QueryParams) bool {
	if !qp.Redirect() {
		return false
	}

	rd, err := s.FormRedirect(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		http.Redirect(w, r, rd, http.StatusMovedPermanently)
	}
	return true
}

// FormRedirect returns the value for the "Location" header for a 301 response.
func (s *Service) FormRedirect(r *http.Request) (string, error) {
	leaderAPIAddr := s.LeaderAPIAddr(r.Context())
	if leaderAPIAddr == "" {
		return "", ErrLeaderNotFound
	}

	rq := r.URL.RawQuery
	if rq != "" {
		rq = fmt.Sprintf("?%s", rq)
	}
	return fmt.Sprintf("%s%s%s", leaderAPIAddr, r.URL.Path, rq), nil
}

// LeaderAPIAddr returns the API address of the leader, as known by this node.
func (s *Service) LeaderAPIAddr(ctx context.Context) string {
	nodeAddr, err := s.store.LeaderAddr(ctx)
	if err != nil {
		return ""
	}
	callCtx, cancel := s.doCtx(ctx)
	defer cancel()
	apiAddr, err := s.cluster.GetNodeAPIAddr(callCtx, nodeAddr)
	if err != nil {
		return ""
	}
	return apiAddr
}

func (s *Service) CheckRequestPerm(r *http.Request, perm v1.Credential_Permission) (b bool) {
	// No auth store set, so no checking required.
	if s.creds == nil {
		return true
	}
	username, password, ok := r.BasicAuth()
	if !ok {
		username = ""
	}

	return s.creds.AA(username, password, perm)
}

func (s *Service) handleReady(w http.ResponseWriter, r *http.Request, params QueryParams) {}

func (s *Service) status() *v1.Status {
	return &v1.Status{
		Os:      s.osStatus(),
		Runtime: s.runtimeStatus(),
		Http:    s.httpStatus(),
		Node:    s.nodeStatus(),
	}
}

func (s *Service) osStatus() *v1.Status_Os {
	o := &v1.Status_Os{
		Pid:      int32(os.Getpid()),
		Ppid:     int32(os.Getppid()),
		PageSize: int32(os.Getpagesize()),
	}
	o.Executable, _ = os.Executable()
	o.Hostname, _ = os.Hostname()
	return o
}

func (s *Service) httpStatus() *v1.Status_HTTP {
	return &v1.Status_HTTP{
		BindAddress: s.ln.Addr().String(),
		EnabledAuth: s.creds != nil,
		Tls:         s.tlsStatus(),
	}
}

func (s *Service) nodeStatus() *v1.Status_Node {
	now := time.Now()
	return &v1.Status_Node{
		StartTime:   timestamppb.New(s.start),
		CurrentTime: timestamppb.New(now),
		Uptime:      durationpb.New(now.Sub(s.start)),
	}
}

func (s *Service) runtimeStatus() *v1.Status_Runtime {
	return &v1.Status_Runtime{
		Os:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		MaxProcs: int32(runtime.GOMAXPROCS(0)),
		NumCpu:   int32(runtime.NumCPU()),
		Version:  runtime.Version(),
	}
}

func (s *Service) tlsStatus() *v1.Status_TLS {
	o := &v1.Status_TLS{
		Enabled: s.tls != nil,
	}
	if s.tls != nil {
		o.ClientAuth = s.tls.ClientAuth.String()
		o.CertFile = s.CertFile
		o.KeyFile = s.KeyFile
		o.CaFile = s.CACertFile
		o.NextProtos = s.tls.NextProtos
	}
	return o
}

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

func makeCredentials(username, password string) *v1.Credential {
	return &v1.Credential{
		Username: username,
		Password: password,
	}
}
