package store

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/apache/arrow/go/v15/arrow/memory"
	"github.com/hashicorp/raft"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/cluster/log"
	"github.com/vinceanalytics/vince/internal/cluster/snapshots"
	"github.com/vinceanalytics/vince/internal/compute"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/index/primary"
	"github.com/vinceanalytics/vince/internal/indexer"
	"github.com/vinceanalytics/vince/internal/lsm"
	"github.com/vinceanalytics/vince/internal/session"
	"github.com/vinceanalytics/vince/internal/tenant"
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

type Store struct {
	open          AtomicBool
	raftDir       string
	snapshotDir   string
	peersPath     string
	peersInfoPath string
	dbPath        string

	restorePath   string
	restoreDoneCh chan struct{}

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

	// Raft changes observer
	leaderObserversMu sync.RWMutex
	leaderObservers   []chan<- struct{}
	observerClose     chan struct{}
	observerDone      chan struct{}
	observerChan      chan raft.Observation
	observer          *raft.Observer

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

	session *session.Tenants
	db      *db.KV
	config  *v1.Config
}

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
		ly:              ly,
		raftDir:         base.Data,
		snapshotDir:     filepath.Join(base.Data, snapshotsDirName),
		peersPath:       filepath.Join(base.Data, peersPath),
		peersInfoPath:   filepath.Join(base.Data, peersInfoPath),
		dbPath:          filepath.Join(base.Data, dbName),
		restoreDoneCh:   make(chan struct{}),
		leaderObservers: make([]chan<- struct{}, 0),
		logger:          slog.Default().With("component", "store"),
		notifyingNodes:  make(map[string]*v1.Server),
		ApplyTimeout:    applyTimeout,
		session:         sess,
		db:              store,
	}, nil
}

func (s *Store) Open() error {
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
	return nil
}

func (s *Store) Data(ctx context.Context, req *v1.Data) error { return nil }

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
