package store

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/apache/arrow/go/v15/arrow/memory"
	"github.com/apache/arrow/go/v15/arrow/util"
	"github.com/docker/go-units"
	"github.com/hashicorp/raft"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/cluster/events"
	"github.com/vinceanalytics/vince/internal/cluster/log"
	"github.com/vinceanalytics/vince/internal/cluster/snapshots"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/index/primary"
	"github.com/vinceanalytics/vince/internal/indexer"
	"google.golang.org/protobuf/proto"
)

// IsStaleRead returns whether a read is stale.
func IsStaleRead(
	leaderlastContact time.Time,
	lastFSMUpdateTime time.Time,
	lastAppendedAtTime time.Time,
	fsmIndex uint64,
	commitIndex uint64,
	freshness int64,
	strict bool,
) bool {
	if freshness == 0 {
		// Freshness not set, so no read can be stale.
		return false
	}
	if time.Since(leaderlastContact).Nanoseconds() > freshness {
		// The Leader has not been in contact witin the freshness window, so
		// the read is stale.
		return true
	}
	if !strict {
		// Strict mode is not enabled, so no further checks are needed.
		return false
	}
	if lastAppendedAtTime.IsZero() {
		// We've yet to be told about any appended log entries, so we
		// assume we're caught up.
		return false
	}
	if fsmIndex == commitIndex {
		// FSM index is the same as the commit index, so we're caught up.
		return false
	}
	// OK, we're not caught up. So was the log that last updated our local FSM
	// appended by the Leader to its log within the freshness window?
	return lastFSMUpdateTime.Sub(lastAppendedAtTime).Nanoseconds() > freshness
}

// IsNewNode returns whether a node using raftDir would be a brand-new node.
// It also means that the window for this node joining a different cluster has passed.
func IsNewNode(raftDir string) bool {
	// If there is any pre-existing Raft state, then this node
	// has already been created.
	return !pathExists(filepath.Join(raftDir, raftDBPath))
}

// HasData returns true if the given dir indicates that at least one FSM entry
// has been committed to the log. This is true if there are any snapshots, or
// if there are any entries in the log of raft.LogCommand type. This function
// will block if the Bolt database is already open.
func HasData(dir string) (bool, error) {
	if !dirExists(dir) {
		return false, nil
	}
	sstr, err := snapshots.New(filepath.Join(dir, snapshotsDirName), 2)
	if err != nil {
		return false, err
	}
	snaps, err := sstr.List()
	if err != nil {
		return false, err
	}
	if len(snaps) > 0 {
		return true, nil
	}
	logs, err := log.New(filepath.Join(dir, raftDBPath), false)
	if err != nil {
		return false, err
	}
	defer logs.Close()
	h, err := logs.HasCommand()
	if err != nil {
		return false, err
	}
	return h, nil
}

// RecoverNode is used to manually force a new configuration, in the event that
// quorum cannot be restored. This borrows heavily from RecoverCluster functionality
// of the Hashicorp Raft library, but has been customized for rqlite use.
func RecoverNode(dataDir string, logs raft.LogStore, stable *log.Log,
	snaps raft.SnapshotStore, tn raft.Transport, conf raft.Configuration) error {
	logger := slog.Default().With("component", "recovery")

	// Sanity check the Raft peer configuration.
	if err := checkRaftConfiguration(conf); err != nil {
		return err
	}

	// Get a path to a temporary file to use for a temporary database.
	tmpDBPath := filepath.Join(dataDir, "recovery")
	defer os.RemoveAll(tmpDBPath)

	// Attempt to restore any latest snapshot.
	var (
		snapshotIndex uint64
		snapshotTerm  uint64
	)

	foundSnaps, err := snaps.List()
	if err != nil {
		return fmt.Errorf("failed to list snapshots: %s", err)
	}
	logger.Info("recovery detected", "snapshots", len(foundSnaps))
	if len(foundSnaps) > 0 {
		if err := func() error {
			snapID := foundSnaps[0].ID
			_, rc, err := snaps.Open(snapID)
			if err != nil {
				return fmt.Errorf("failed to open snapshot %s: %s", snapID, err)
			}
			defer rc.Close()
			bdb, err := db.OpenBadger(tmpDBPath)
			if err != nil {
				return fmt.Errorf("failed to copy snapshot %s to temporary database: %s", snapID, err)
			}
			err = bdb.Load(rc, runtime.NumCPU())
			if err != nil {
				return fmt.Errorf("failed to load snapshot %s to temporary database: %s", snapID, err)
			}
			bdb.Close()
			snapshotIndex = foundSnaps[0].Index
			snapshotTerm = foundSnaps[0].Term
			return nil
		}(); err != nil {
			return err
		}
	}

	// Now, open the database so we can replay any outstanding Raft log entries.
	rdb, err := db.NewKV(tmpDBPath)
	if err != nil {
		return fmt.Errorf("failed to open temporary database: %s", err)
	}
	defer rdb.Close()

	// The snapshot information is the best known end point for the data
	// until we play back the Raft log entries.
	lastIndex := snapshotIndex
	lastTerm := snapshotTerm

	// Apply any Raft log entries past the snapshot.
	lastLogIndex, err := logs.LastIndex()
	if err != nil {
		return fmt.Errorf("failed to find last log: %v", err)
	}
	logger.Info("Applying raft log", "lastIndex", lastIndex, "lastLogIndex", lastLogIndex)
	reco := newRecovery(logger, rdb)
	var eve v1.Data
	for index := snapshotIndex + 1; index <= lastLogIndex; index++ {
		var entry raft.Log
		if err = logs.GetLog(index, &entry); err != nil {
			return fmt.Errorf("failed to get log at index %d: %v", index, err)
		}
		if entry.Type == raft.LogCommand {
			err = proto.Unmarshal(entry.Data, &eve)
			if err != nil {
				return fmt.Errorf("failed to decode raft command %v", err)
			}
			err = reco.write(&eve)
			if err != nil {
				return err
			}
		}
		lastIndex = entry.Index
		lastTerm = entry.Term
	}
	err = reco.flush()
	if err != nil {
		return err
	}
	tmpDBFD, err := db.OpenBadger(tmpDBPath)
	if err != nil {
		return fmt.Errorf("failed to open temporary database file: %s", err)
	}
	fsmSnapshot := snapshots.NewBadger(tmpDBFD) // tmpDBPath contains full state now.
	sink, err := snaps.Create(1, lastIndex, lastTerm, conf, 1, tn)
	if err != nil {
		return fmt.Errorf("failed to create snapshot: %v", err)
	}
	if err = fsmSnapshot.Persist(sink); err != nil {
		return fmt.Errorf("failed to persist snapshot: %v", err)
	}
	if err = sink.Close(); err != nil {
		return fmt.Errorf("failed to finalize snapshot: %v", err)
	}
	logger.Info("recovery snapshot created successfully", "db", tmpDBPath)

	// Compact the log so that we don't get bad interference from any
	// configuration change log entries that might be there.
	firstLogIndex, err := logs.FirstIndex()
	if err != nil {
		return fmt.Errorf("failed to get first log index: %v", err)
	}
	if err := logs.DeleteRange(firstLogIndex, lastLogIndex); err != nil {
		return fmt.Errorf("log compaction failed: %v", err)
	}

	// Erase record of previous updating of Applied Index too.
	if err := stable.SetAppliedIndex(0); err != nil {
		return fmt.Errorf("failed to zero applied index: %v", err)
	}
	return nil
}

// checkRaftConfiguration tests a cluster membership configuration for common
// errors.
func checkRaftConfiguration(configuration raft.Configuration) error {
	idSet := make(map[raft.ServerID]bool)
	addressSet := make(map[raft.ServerAddress]bool)
	var voters int
	for _, server := range configuration.Servers {
		if server.ID == "" {
			return fmt.Errorf("empty ID in configuration: %v", configuration)
		}
		if server.Address == "" {
			return fmt.Errorf("empty address in configuration: %v", server)
		}
		if strings.Contains(string(server.Address), "://") {
			return fmt.Errorf("protocol specified in address: %v", server.Address)
		}
		_, _, err := net.SplitHostPort(string(server.Address))
		if err != nil {
			return fmt.Errorf("invalid address in configuration: %v", server.Address)
		}
		if idSet[server.ID] {
			return fmt.Errorf("found duplicate ID in configuration: %v", server.ID)
		}
		idSet[server.ID] = true
		if addressSet[server.Address] {
			return fmt.Errorf("found duplicate address in configuration: %v", server.Address)
		}
		addressSet[server.Address] = true
		if server.Suffrage == raft.Voter {
			voters++
		}
	}
	if voters == 0 {
		return fmt.Errorf("need at least one voter in configuration: %v", configuration)
	}
	return nil
}

const (
	maxRows = 4 << 20
)

type recoveryInstance struct {
	mem     memory.Allocator
	build   map[string]*events.Builder
	store   *db.Store
	primary *primary.PrimaryIndex
	index   indexer.ArrowIndexer
	log     *slog.Logger
}

func newRecovery(log *slog.Logger, store db.Storage) *recoveryInstance {
	mem := memory.NewGoAllocator()
	return &recoveryInstance{
		mem:     mem,
		build:   make(map[string]*events.Builder),
		store:   db.NewStore(store, mem, 0),
		primary: primary.Empty(store),
		log:     log,
	}
}

func (r *recoveryInstance) write(e *v1.Data) error {
	b, ok := r.build[e.TenantId]
	if !ok {
		b = events.New(r.mem)
		r.build[e.TenantId] = b
	}
	b.WriteData(e)
	if b.Len() >= maxRows {
		err := r.save(e.TenantId, b)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *recoveryInstance) Release() {
	for _, b := range r.build {
		b.Release()
	}
}

func (r *recoveryInstance) flush() error {
	for tenant, b := range r.build {
		err := r.save(tenant, b)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *recoveryInstance) save(tenant string, b *events.Builder) error {
	a := b.NewRecord()
	defer a.Release()
	r.log.Info("Building index",
		"tenant", tenant,
		"rows", b.Len())
	full, err := r.index.Index(a)
	if err != nil {
		return err
	}
	r.log.Info("Saving record and index",
		"tenant", tenant,
		"record_size", units.BytesSize(float64(util.TotalRecordSize(a))),
		"index_size", units.BytesSize(float64(full.Size())),
	)
	g, err := r.store.Save(tenant, a, full)
	if err != nil {
		return err
	}
	r.log.Info("Saved record and index",
		"tenant", tenant,
		"total_size", units.BytesSize(float64(g.Size)),
		"block_id", g.Id,
	)
	r.log.Info("Updating primary index",
		"tenant", tenant)
	return r.primary.Add(tenant, g)
}
