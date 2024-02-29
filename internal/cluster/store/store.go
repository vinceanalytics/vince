package store

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/apache/arrow/go/v15/arrow/memory"
	"github.com/hashicorp/raft"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/cluster/events"
	"github.com/vinceanalytics/vince/internal/cluster/log"
	"github.com/vinceanalytics/vince/internal/cluster/snapshots"
	"github.com/vinceanalytics/vince/internal/compute"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/index/primary"
	"github.com/vinceanalytics/vince/internal/indexer"
	"github.com/vinceanalytics/vince/internal/lsm"
	"github.com/vinceanalytics/vince/internal/session"
	"github.com/vinceanalytics/vince/internal/tenant"
	"google.golang.org/protobuf/proto"
)

var (
	// ErrNotOpen is returned when a Store is not open.
	ErrNotOpen = errors.New("store not open")

	// ErrOpen is returned when a Store is already open.
	ErrOpen = errors.New("store already open")

	// ErrNotReady is returned when a Store is not ready to accept requests.
	ErrNotReady = errors.New("store not ready")

	// ErrNotLeader is returned when a node attempts to execute a leader-only
	// operation.
	ErrNotLeader = errors.New("not leader")

	// ErrNotSingleNode is returned when a node attempts to execute a single-node
	// only operation.
	ErrNotSingleNode = errors.New("not single-node")

	// ErrStaleRead is returned if the executing the query would violate the
	// requested freshness.
	ErrStaleRead = errors.New("stale read")

	// ErrOpenTimeout is returned when the Store does not apply its initial
	// logs within the specified time.
	ErrOpenTimeout = errors.New("timeout waiting for initial logs application")

	// ErrWaitForRemovalTimeout is returned when the Store does not confirm removal
	// of a node within the specified time.
	ErrWaitForRemovalTimeout = errors.New("timeout waiting for node removal confirmation")

	// ErrWaitForLeaderTimeout is returned when the Store cannot determine the leader
	// within the specified time.
	ErrWaitForLeaderTimeout = errors.New("timeout waiting for leader")

	// ErrInvalidBackupFormat is returned when the requested backup format
	// is not valid.
	ErrInvalidBackupFormat = errors.New("invalid backup format")

	// ErrInvalidVacuumFormat is returned when the requested backup format is not
	// compatible with vacuum.
	ErrInvalidVacuum = errors.New("invalid vacuum")

	// ErrLoadInProgress is returned when a load is already in progress and the
	// requested operation cannot be performed.
	ErrLoadInProgress = errors.New("load in progress")
)

const (
	defaultTimeout = 30 * time.Second

	snapshotsDirName           = "snapshots"
	dbName                     = "vince"
	restoreScratchPattern      = "rqlite-restore-*"
	bootScatchPattern          = "rqlite-boot-*"
	backupScatchPattern        = "rqlite-backup-*"
	vacuumScatchPattern        = "rqlite-vacuum-*"
	raftDBPath                 = "raftdb" // Changing this will break backwards compatibility.
	peersPath                  = "raft/peers.json"
	peersInfoPath              = "raft/peers.info"
	retainSnapshotCount        = 1
	applyTimeout               = 10 * time.Second
	openTimeout                = 120 * time.Second
	leaderWaitDelay            = 100 * time.Millisecond
	appliedWaitDelay           = 100 * time.Millisecond
	commitEquivalenceDelay     = 50 * time.Millisecond
	appliedIndexUpdateInterval = 5 * time.Second
	connectionPoolCount        = 5
	connectionTimeout          = 10 * time.Second
	raftLogCacheSize           = 512
	trailingScale              = 1.25
	observerChanLen            = 50
)

type SnapshotStore interface {
	raft.SnapshotStore

	// FullNeeded returns true if a full snapshot is needed.
	FullNeeded() (bool, error)

	// SetFullNeeded explicitly sets that a full snapshot is needed.
	SetFullNeeded() error
}

// ClusterState defines the possible Raft states the current node can be in
type ClusterState int

// Represents the Raft cluster states
const (
	Leader ClusterState = iota
	Follower
	Candidate
	Shutdown
	Unknown
)

type Database interface {
	Data(ctx context.Context, req *v1.Data) error
	Realtime(ctx context.Context, req *v1.Realtime_Request) (*v1.Realtime_Response, error)
	Aggregate(ctx context.Context, req *v1.Aggregate_Request) (*v1.Aggregate_Response, error)
	Timeseries(ctx context.Context, req *v1.Timeseries_Request) (*v1.Timeseries_Response, error)
	Breakdown(ctx context.Context, req *v1.BreakDown_Request) (*v1.BreakDown_Response, error)
	Load(ctx context.Context, req *v1.Load_Request) error
}

// Storage is the interface the Raft-based database must implement.
type Storage interface {
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
type Store struct {
	open          AtomicBool
	raftDir       string
	snapshotDir   string
	peersPath     string
	peersInfoPath string
	dbPath        string

	raft   *raft.Raft // The consensus mechanism.
	ly     Layer
	raftTn *NodeTransport

	// Channels that must be closed for the Store to be considered ready.
	readyChans             []<-chan struct{}
	numClosedReadyChannels int
	readyChansMu           sync.Mutex

	// Channels for WAL-size triggered snapshotting
	snapshotWClose chan struct{}
	snapshotWDone  chan struct{}
	snapshotCAS    CheckAndSet

	// Latest log entry index actually reflected by the FSM. Due to Raft code
	// this value is not updated after a Snapshot-restore.
	fsmIdx        atomic.Uint64
	fsmUpdateTime *AtomicTime // This is node-local time.

	// appendedAtTimeis the Leader's clock time when that Leader appended the log entry.
	// The Leader that actually appended the log entry is not necessarily the current Leader.
	appendedAtTime AtomicTime

	raftLog       raft.LogStore    // Persistent log store.
	raftStable    raft.StableStore // Persistent k-v store.
	snapshotStore *snapshots.Store // Snapshot store.
	boltStore     *log.Log

	firstIdxOnOpen       uint64    // First index on log when Store opens.
	lastIdxOnOpen        uint64    // Last index on log when Store opens.
	lastCommandIdxOnOpen uint64    // Last command index before applied index when Store opens.
	lastAppliedIdxOnOpen uint64    // Last applied index on log when Store opens.
	firstLogAppliedT     time.Time // Time first log is applied
	appliedOnOpen        uint64    // Number of logs applied at open.
	openT                time.Time // Timestamp when Store opens.

	logger         *slog.Logger
	logIncremental bool

	notifyMu        sync.Mutex
	BootstrapExpect int
	bootstrapped    bool
	notifyingNodes  map[string]*v1.Server

	ShutdownOnRemove         bool
	SnapshotThreshold        uint64
	SnapshotThresholdWALSize uint64
	SnapshotInterval         time.Duration
	LeaderLeaseTimeout       time.Duration
	HeartbeatTimeout         time.Duration
	ElectionTimeout          time.Duration
	ApplyTimeout             time.Duration
	RaftLogLevel             string
	NoFreeListSync           bool
	AutoVacInterval          time.Duration

	session *session.Session
	db      *db.KV
	config  *v1.Config
}

var _ Storage = (*Store)(nil)

func NewStore(base *v1.Config, ly Layer) (*Store, error) {
	store, err := db.NewKV(base.Data)
	if err != nil {
		return nil, err
	}
	tenants := tenant.NewTenants(base)
	alloc := memory.NewGoAllocator()
	pidx, err := primary.NewPrimary(store)
	if err != nil {
		return nil, err
	}
	idx := indexer.New()
	sess := session.New(alloc, tenants, store, idx, pidx,
		lsm.WithTTL(
			base.RetentionPeriod.AsDuration(),
		),
		lsm.WithCompactSize(
			uint64(base.GranuleSize),
		),
	)

	return &Store{
		ly:             ly,
		raftDir:        base.Data,
		snapshotDir:    filepath.Join(base.Data, snapshotsDirName),
		peersPath:      filepath.Join(base.Data, peersPath),
		peersInfoPath:  filepath.Join(base.Data, peersInfoPath),
		dbPath:         filepath.Join(base.Data, dbName),
		logger:         slog.Default().With("component", "store"),
		notifyingNodes: make(map[string]*v1.Server),
		ApplyTimeout:   applyTimeout,
		session:        sess,
		db:             store,
	}, nil
}

func (s *Store) Open(ctx context.Context) error {
	if s.open.Is() {
		return ErrOpen
	}
	s.openT = time.Now()
	s.logger.Info("Opening store", "nodeId", s.config.NodeId, "listening", s.ly.Addr().String())

	_, err := os.Stat(s.config.Data)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(s.config.Data, 0755)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	// Create Raft-compatible network layer.
	nt := raft.NewNetworkTransport(NewTransport(s.ly), connectionPoolCount, connectionTimeout, nil)
	s.raftTn = NewNodeTransport(nt)

	s.snapshotStore, err = snapshots.New(s.snapshotDir, 2)
	if err != nil {
		return err
	}
	s.boltStore, err = log.New(filepath.Join(s.raftDir, raftDBPath), false)
	if err != nil {
		return err
	}
	s.raftStable = s.boltStore
	s.raftLog, err = raft.NewLogCache(raftLogCacheSize, s.boltStore)
	if err != nil {
		return err
	}
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(s.config.NodeId)

	// Request to recover node?
	if pathExists(s.peersPath) {
		s.logger.Info("attempting node recovery ", "peerPath", s.peersPath)
		config, err := raft.ReadConfigJSON(s.peersPath)
		if err != nil {
			return fmt.Errorf("failed to read peers file: %s", err.Error())
		}
		if err = RecoverNode(s.raftDir, s.raftLog, s.boltStore, s.snapshotStore, s.raftTn, config); err != nil {
			return fmt.Errorf("failed to recover node: %s", err.Error())
		}
		if err := os.Rename(s.peersPath, s.peersInfoPath); err != nil {
			return fmt.Errorf("failed to move %s after recovery: %s", s.peersPath, err.Error())
		}
		s.logger.Info("node recovered successfully", "peerPath", s.peersPath)
	}

	// Get some info about the log, before any more entries are committed.
	if err := s.setLogInfo(); err != nil {
		return fmt.Errorf("set log info: %s", err)
	}
	s.logger.Info("Setup log info",
		"firstIdxOnOpen", s.firstIdxOnOpen,
		"lastIdxOnOpen", s.lastIdxOnOpen,
		"lastAppliedIdxOnOpen", s.lastAppliedIdxOnOpen,
		"lastCommandIdxOnOpen", s.lastCommandIdxOnOpen,
	)
	s.db, err = db.NewKV(s.dbPath)
	if err != nil {
		return fmt.Errorf("failed to create on-disk database: %s", err)
	}

	// Instantiate the Raft system.
	ra, err := raft.NewRaft(config, NewFSM(s), s.raftLog, s.raftStable, s.snapshotStore, s.raftTn)
	if err != nil {
		return fmt.Errorf("creating the raft system failed: %s", err)
	}
	s.raft = ra
	s.session.Start(ctx)
	return nil
}

func (s *Store) Status() (*v1.Status_Store, error) {
	return &v1.Status_Store{}, nil
}

// ReadFrom reads data from r, and loads it into the database, bypassing Raft consensus.
// Once the data is loaded, a snapshot is triggered, which then results in a system as
// if the data had been loaded through Raft consensus.
func (s *Store) ReadFrom(ctx context.Context, r io.Reader) (int64, error) {
	// Check the constraints.
	if s.raft.State() != raft.Leader {
		return 0, ErrNotLeader
	}
	nodes, err := s.Nodes(ctx)
	if err != nil {
		return 0, err
	}
	if len(nodes.Items) != 1 {
		return 0, ErrNotSingleNode
	}

	err = s.db.DB.Load(r, runtime.NumCPU())
	if err != nil {
		return 0, err
	}

	// Raft won't snapshot unless there is at least one unsnappshotted log entry,
	// so prep that now before we do anything destructive.
	n, err := s.sendData(ctx, nil)
	if err != nil {
		return 0, err
	}

	if err := s.Snapshot(1); err != nil {
		return 0, err
	}
	return int64(n), nil
}

func (s *Store) Backup(ctx context.Context, br *v1.Backup_Request, dst io.Writer) error {
	if !s.open.Is() {
		return ErrNotOpen
	}
	if br.Leader && s.raft.State() != raft.Leader {
		return ErrNotLeader
	}
	// Snapshot to ensure the main SQLite file has all the latest data.
	if err := s.Snapshot(0); err != nil {
		if err != raft.ErrNothingNewToSnapshot &&
			!strings.Contains(err.Error(), "wait until the configuration entry at") {
			return fmt.Errorf("pre-backup snapshot failed: %s", err.Error())
		}
	}
	if err := s.snapshotCAS.Begin(); err != nil {
		return err
	}
	defer s.snapshotCAS.End()
	_, err := s.db.DB.Backup(dst, 0)
	return err
}

// Snapshot performs a snapshot, leaving n trailing logs behind. If n
// is greater than zero, that many logs are left in the log after
// snapshotting. If n is zero, then the number set at Store creation is used.
// Finally, once this function returns, the trailing log configuration value
// is reset to the value set at Store creation.
func (s *Store) Snapshot(n uint64) (retError error) {

	if n > 0 {
		cfg := s.raft.ReloadableConfig()
		defer func() {
			if err := s.raft.ReloadConfig(cfg); err != nil {
				s.logger.Error("failed to reload Raft config", "err", err)
			}
		}()
		cfg.TrailingLogs = n
		if err := s.raft.ReloadConfig(cfg); err != nil {
			return fmt.Errorf("failed to reload Raft config: %s", err.Error())
		}
	}
	if err := s.raft.Snapshot().Error(); err != nil {
		if strings.Contains(err.Error(), ErrLoadInProgress.Error()) {
			return ErrLoadInProgress
		}
		return err
	}

	return nil
}

func (s *Store) Data(ctx context.Context, req *v1.Data) error {
	if !s.open.Is() {
		return ErrNotOpen
	}
	if s.raft.State() != raft.Leader {
		return ErrNotLeader
	}
	if !s.Ready(ctx) {
		return ErrNotReady
	}
	_, err := s.sendData(ctx, req)
	return err
}

func (s *Store) sendData(ctx context.Context, req *v1.Data) (uint64, error) {
	var b []byte
	var err error
	if req != nil {
		b, err = proto.Marshal(req)
		if err != nil {
			return 0, err
		}
	}

	af := s.raft.Apply(b, s.ApplyTimeout)
	if af.Error() != nil {
		if af.Error() == raft.ErrNotLeader {
			return 0, ErrNotLeader
		}
		return 0, af.Error()
	}
	r := af.Response()
	if r != nil {
		return 0, r.(error)
	}
	return af.Index(), nil
}

// Ready returns true if the store is ready to serve requests. Ready is
// defined as having no open channels registered via RegisterReadyChannel
// and having a Leader.
func (s *Store) Ready(ctx context.Context) bool {
	l, _ := s.LeaderAddr(ctx)
	if l == "" {
		return false
	}

	return func() bool {
		s.readyChansMu.Lock()
		defer s.readyChansMu.Unlock()
		if s.numClosedReadyChannels != len(s.readyChans) {
			return false
		}
		s.readyChans = nil
		s.numClosedReadyChannels = 0
		return true
	}()
}

// Committed blocks until the local commit index is greater than or
// equal to the Leader index, as checked when the function is called.
// It returns the committed index. If the Leader index is 0, then the
// system waits until the commit index is at least 1.
func (s *Store) Committed(ctx context.Context) (uint64, error) {
	lci, err := s.LeaderCommitIndex()
	if err != nil {
		return lci, err
	}
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	return lci, s.WaitForCommitIndex(ctx, max(1, lci))
}

// LeaderCommitIndex returns the Raft leader commit index, as indicated
// by the latest AppendEntries RPC. If this node is the Leader then the
// commit index is returned directly from the Raft object.
func (s *Store) LeaderCommitIndex() (uint64, error) {
	if s.raft.State() == raft.Leader {
		return s.raft.CommitIndex(), nil
	}
	return s.raftTn.LeaderCommitIndex(), nil
}

// Nodes returns the slice of nodes in the cluster, sorted by ID ascending.
func (s *Store) Nodes(ctx context.Context) (*v1.Server_List, error) {
	if !s.open.Is() {
		return nil, ErrNotOpen
	}

	f := s.raft.GetConfiguration()
	if f.Error() != nil {
		return nil, f.Error()
	}

	rs := f.Configuration().Servers
	servers := make([]*v1.Server, len(rs))
	for i := range rs {
		servers[i] = &v1.Server{
			Id:       string(rs[i].ID),
			Addr:     string(rs[i].Address),
			Suffrage: v1.Server_Suffrage(v1.Server_Suffrage_value[rs[i].Suffrage.String()]),
		}
	}
	sort.Slice(servers, func(i, j int) bool {
		return servers[i].Id < servers[j].Id
	})
	return &v1.Server_List{Items: servers}, nil
}

// Close closes the store. If wait is true, waits for a graceful shutdown.
func (s *Store) Close(wait bool) (retErr error) {
	defer func() {
		if retErr == nil {
			s.logger.Info("store closed ", "nodeId", s.config.NodeId, "listen_address", s.ly.Addr().String())
			s.open.Unset()
		}
	}()
	if !s.open.Is() {
		// Protect against closing already-closed resource, such as channels.
		return nil
	}

	close(s.snapshotWClose)
	<-s.snapshotWDone

	f := s.raft.Shutdown()
	if wait {
		if f.Error() != nil {
			return f.Error()
		}
	}
	if err := s.raftTn.Close(); err != nil {
		return err
	}

	s.session.Close()

	// Only shutdown Bolt and SQLite when Raft is done.
	if err := s.db.Close(); err != nil {
		return err
	}
	if err := s.boltStore.Close(); err != nil {
		return err
	}
	return nil
}

// WaitForAppliedFSM waits until the currently applied logs (at the time this
// function is called) are actually reflected by the FSM, or the timeout expires.
func (s *Store) WaitForAppliedFSM(timeout time.Duration) (uint64, error) {
	if timeout == 0 {
		return 0, nil
	}
	return s.WaitForFSMIndex(s.raft.AppliedIndex(), timeout)
}

// WaitForApplied waits for all Raft log entries to be applied to the
// underlying database.
func (s *Store) WaitForAllApplied(timeout time.Duration) error {
	if timeout == 0 {
		return nil
	}
	return s.WaitForAppliedIndex(s.raft.LastIndex(), timeout)
}

// WaitForAppliedIndex blocks until a given log index has been applied,
// or the timeout expires.
func (s *Store) WaitForAppliedIndex(idx uint64, timeout time.Duration) error {
	tck := time.NewTicker(appliedWaitDelay)
	defer tck.Stop()
	tmr := time.NewTimer(timeout)
	defer tmr.Stop()

	for {
		select {
		case <-tck.C:
			if s.raft.AppliedIndex() >= idx {
				return nil
			}
		case <-tmr.C:
			return fmt.Errorf("timeout expired")
		}
	}
}

// WaitForFSMIndex blocks until a given log index has been applied to our
// state machine or the timeout expires.
func (s *Store) WaitForFSMIndex(idx uint64, timeout time.Duration) (uint64, error) {
	tck := time.NewTicker(appliedWaitDelay)
	defer tck.Stop()
	tmr := time.NewTimer(timeout)
	defer tmr.Stop()
	for {
		select {
		case <-tck.C:
			if fsmIdx := s.fsmIdx.Load(); fsmIdx >= idx {
				return fsmIdx, nil
			}
		case <-tmr.C:
			return 0, fmt.Errorf("timeout expired")
		}
	}
}

// WaitForCommitIndex blocks until the local Raft commit index is equal to
// or greater the given index, or the timeout expires.
func (s *Store) WaitForCommitIndex(ctx context.Context, idx uint64) error {
	tck := time.NewTicker(commitEquivalenceDelay)
	defer tck.Stop()
	checkFn := func() bool {
		return s.raft.CommitIndex() >= idx
	}

	// Try the fast path.
	if checkFn() {
		return nil
	}
	for {
		select {
		case <-tck.C:
			if checkFn() {
				return nil
			}
		case <-ctx.Done():
			return fmt.Errorf("timeout expired")
		}
	}
}

// IsLeader is used to determine if the current node is cluster leader
func (s *Store) IsLeader() bool {
	return s.raft.State() == raft.Leader
}

// HasLeader returns true if the cluster has a leader, false otherwise.
func (s *Store) HasLeader() bool {
	return s.raft.Leader() != ""
}

// IsVoter returns true if the current node is a voter in the cluster. If there
// is no reference to the current node in the current cluster configuration then
// false will also be returned.
func (s *Store) IsVoter() (bool, error) {
	cfg := s.raft.GetConfiguration()
	if err := cfg.Error(); err != nil {
		return false, err
	}
	for _, srv := range cfg.Configuration().Servers {
		if srv.ID == raft.ServerID(s.config.NodeId) {
			return srv.Suffrage == raft.Voter, nil
		}
	}
	return false, nil
}

// State returns the current node's Raft state
func (s *Store) State() ClusterState {
	state := s.raft.State()
	switch state {
	case raft.Leader:
		return Leader
	case raft.Candidate:
		return Candidate
	case raft.Follower:
		return Follower
	case raft.Shutdown:
		return Shutdown
	default:
		return Unknown
	}
}

// Path returns the path to the store's storage directory.
func (s *Store) Path() string {
	return s.raftDir
}

// Addr returns the address of the store.
func (s *Store) Addr() string {
	if !s.open.Is() {
		return ""
	}
	return string(s.raftTn.LocalAddr())
}

// ID returns the Raft ID of the store.
func (s *Store) ID() string {
	return s.config.NodeId
}

// LeaderAddr returns the address of the current leader. Returns a
// blank string if there is no leader or if the Store is not open.
func (s *Store) LeaderAddr(_ context.Context) (string, error) {
	if !s.open.Is() {
		return "", nil
	}
	addr, _ := s.raft.LeaderWithID()
	return string(addr), nil
}

func (s *Store) Realtime(ctx context.Context, req *v1.Realtime_Request) (*v1.Realtime_Response, error) {
	if s.raft.State() != raft.Leader {
		return nil, ErrNotLeader
	}
	return compute.Realtime(ctx, s.session, req)
}

func (s *Store) Aggregate(ctx context.Context, req *v1.Aggregate_Request) (*v1.Aggregate_Response, error) {
	if s.raft.State() != raft.Leader {
		return nil, ErrNotLeader
	}
	return compute.Aggregate(ctx, s.session, req)
}

func (s *Store) Timeseries(ctx context.Context, req *v1.Timeseries_Request) (*v1.Timeseries_Response, error) {
	if s.raft.State() != raft.Leader {
		return nil, ErrNotLeader
	}
	return compute.Timeseries(ctx, s.session, req)
}

func (s *Store) Breakdown(ctx context.Context, req *v1.BreakDown_Request) (*v1.BreakDown_Response, error) {
	if s.raft.State() != raft.Leader {
		return nil, ErrNotLeader
	}
	return compute.Breakdown(ctx, s.session, req)
}

func (s *Store) Load(ctx context.Context, req *v1.Load_Request) error { return nil }

func (s *Store) fsmApply(l *raft.Log) interface{} {
	if s.firstLogAppliedT.IsZero() {
		s.firstLogAppliedT = time.Now()
		s.logger.Info("first log applied since node start, log", "index", l.Index)
	}
	if len(l.Data) == 0 {
		return nil
	}
	e := events.One()
	err := proto.Unmarshal(l.Data, e)
	if err != nil {
		return err
	}
	s.session.Append(e)
	return nil
}

func (s *Store) fsmSnapshot() (raft.FSMSnapshot, error) {
	s.session.Persist()
	return snapshots.NewBadger(s.db.DB), nil
}

func (s *Store) fsmRestore(w io.ReadCloser) error {
	s.logger.Info("initiating node restore", "nodeId", s.config.NodeId)
	startT := time.Now()

	err := snapshots.NewBadger(s.db.DB).Restore(w)
	if err != nil {
		return err
	}
	meta, err := s.snapshotStore.List()
	if err != nil {
		return fmt.Errorf("failed to get latest snapshot index post restore: %s", err)
	}
	idx := meta[len(meta)-1].Index
	if err := s.boltStore.SetAppliedIndex(idx); err != nil {
		return fmt.Errorf("failed to set applied index: %s", err)
	}
	s.fsmIdx.Store(idx)
	elapsed := time.Since(startT)
	s.logger.Info("node restored", "elapsed", elapsed)
	return nil
}

// setLogInfo records some key indexes about the log.
func (s *Store) setLogInfo() error {
	var err error
	s.firstIdxOnOpen, err = s.boltStore.FirstIndex()
	if err != nil {
		return fmt.Errorf("failed to get last index: %s", err)
	}
	s.lastAppliedIdxOnOpen, err = s.boltStore.GetAppliedIndex()
	if err != nil {
		return fmt.Errorf("failed to get last applied index: %s", err)
	}
	s.lastIdxOnOpen, err = s.boltStore.LastIndex()
	if err != nil {
		return fmt.Errorf("failed to get last index: %s", err)
	}
	s.lastCommandIdxOnOpen, err = s.boltStore.LastCommandIndex(s.firstIdxOnOpen, s.lastAppliedIdxOnOpen)
	if err != nil {
		return fmt.Errorf("failed to get last command index: %s", err)
	}
	return nil
}

// Remove removes a node from the store.
func (s *Store) Remove(ctx context.Context, rn *v1.RemoveNode_Request) error {
	if !s.open.Is() {
		return ErrNotOpen
	}
	id := rn.Id
	s.logger.Info("received request to remove node ", "nodeId", id)
	if err := s.remove(id); err != nil {
		return err
	}
	s.logger.Info("node removed successfully", "nodeId", id)
	return nil
}

// remove removes the node, with the given ID, from the cluster.
func (s *Store) remove(id string) error {
	f := s.raft.RemoveServer(raft.ServerID(id), 0, 0)
	if f.Error() != nil && f.Error() == raft.ErrNotLeader {
		return ErrNotLeader
	}
	return f.Error()
}

// pathExists returns true if the given path exists.
func pathExists(p string) bool {
	if _, err := os.Lstat(p); err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

// pathExistsWithData returns true if the given path exists and has data.
func pathExistsWithData(p string) bool {
	if !pathExists(p) {
		return false
	}
	if size, err := fileSize(p); err != nil || size == 0 {
		return false
	}
	return true
}

func dirExists(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && stat.IsDir()
}

func fileSize(path string) (int64, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return stat.Size(), nil
}

// fileSizeExists returns the size of the given file, or 0 if the file does not
// exist. Any other error is returned.
func fileSizeExists(path string) (int64, error) {
	if !pathExists(path) {
		return 0, nil
	}
	return fileSize(path)
}

// dirSize returns the total size of all files in the given directory
func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			// If the file doesn't exist, we can ignore it. Snapshot files might
			// disappear during walking.
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

func resolvableAddress(addr string) (string, error) {
	h, _, err := net.SplitHostPort(addr)
	if err != nil {
		// Just try the given address directly.
		h = addr
	}
	_, err = net.LookupHost(h)
	return h, err
}
