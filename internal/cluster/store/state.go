package store

import (
	"fmt"
	"net"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/raft"
	"github.com/vinceanalytics/vince/internal/cluster/log"
	"github.com/vinceanalytics/vince/internal/cluster/snapshots"
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
