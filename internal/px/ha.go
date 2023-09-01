package px

import (
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/hashicorp/raft"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/raft/v1"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func AppendEntriesRequest(r *v1.AppendEntriesRequest) *raft.AppendEntriesRequest {
	entries := make([]*raft.Log, len(r.Entries))
	for i := range entries {
		entries[i] = RaftLog(r.Entries[i])
	}
	return &raft.AppendEntriesRequest{
		RPCHeader:         Header(r.Header),
		Term:              r.Term,
		PrevLogEntry:      r.PrevLogEntry,
		PrevLogTerm:       r.PrevLogTerm,
		Entries:           entries,
		LeaderCommitIndex: r.LeaderCommitIndex,
	}
}

func AppendEntriesRequestFrom(r *raft.AppendEntriesRequest) *v1.AppendEntriesRequest {
	entries := make([]*v1.RaftLog, len(r.Entries))
	for i := range entries {
		entries[i] = RaftLogFrom(r.Entries[i])
	}
	return &v1.AppendEntriesRequest{
		Header:            HeaderFrom(&r.RPCHeader),
		Term:              r.Term,
		PrevLogEntry:      r.PrevLogEntry,
		PrevLogTerm:       r.PrevLogTerm,
		Entries:           entries,
		LeaderCommitIndex: r.LeaderCommitIndex,
	}
}

func AppendEntriesResponse(r *v1.AppendEntriesResponse) *raft.AppendEntriesResponse {
	return &raft.AppendEntriesResponse{
		RPCHeader:      Header(r.Header),
		Term:           r.Term,
		LastLog:        r.LastLog,
		Success:        r.Success,
		NoRetryBackoff: r.NoRetryBackoff,
	}
}

func AppendEntriesResponseFrom(r *raft.AppendEntriesResponse) *v1.AppendEntriesResponse {
	return &v1.AppendEntriesResponse{
		Header:         HeaderFrom(&r.RPCHeader),
		Term:           r.Term,
		LastLog:        r.LastLog,
		Success:        r.Success,
		NoRetryBackoff: r.NoRetryBackoff,
	}
}

func RequestVoteRequest(r *v1.RequestVoteRequest) *raft.RequestVoteRequest {
	return &raft.RequestVoteRequest{
		RPCHeader:          Header(r.Header),
		Term:               r.Term,
		LastLogIndex:       r.LastLogIndex,
		LastLogTerm:        r.LastLogTerm,
		LeadershipTransfer: r.LeadershipTransfer,
	}
}

func RequestVoteRequestFrom(r *raft.RequestVoteRequest) *v1.RequestVoteRequest {
	return &v1.RequestVoteRequest{
		Header:             HeaderFrom(&r.RPCHeader),
		Term:               r.Term,
		LastLogIndex:       r.LastLogIndex,
		LastLogTerm:        r.LastLogTerm,
		LeadershipTransfer: r.LeadershipTransfer,
	}
}

func RequestVoteResponse(r *v1.RequestVoteResponse) *raft.RequestVoteResponse {
	return &raft.RequestVoteResponse{
		RPCHeader: Header(r.Header),
		Term:      r.Term,
		Granted:   r.Granted,
	}
}

func RequestVoteResponseFrom(r *raft.RequestVoteResponse) *v1.RequestVoteResponse {
	return &v1.RequestVoteResponse{
		Header:  HeaderFrom(&r.RPCHeader),
		Term:    r.Term,
		Granted: r.Granted,
	}
}

func InstallSnapshotRequest(r *v1.InstallSnapshotRequest) *raft.InstallSnapshotRequest {
	return &raft.InstallSnapshotRequest{
		RPCHeader:          Header(r.Header),
		Term:               r.Term,
		Leader:             r.Leader,
		LastLogIndex:       r.LastLogIndex,
		LastLogTerm:        r.LastLogTerm,
		Configuration:      r.Configuration,
		ConfigurationIndex: r.ConfigurationIndex,
		Size:               r.Size,
	}
}

func InstallSnapshotRequestFrom(r *raft.InstallSnapshotRequest) *v1.InstallSnapshotRequest {
	return &v1.InstallSnapshotRequest{
		Header:             HeaderFrom(&r.RPCHeader),
		Term:               r.Term,
		Leader:             r.Leader,
		LastLogIndex:       r.LastLogIndex,
		LastLogTerm:        r.LastLogTerm,
		Configuration:      r.Configuration,
		ConfigurationIndex: r.ConfigurationIndex,
		Size:               r.Size,
	}
}

func InstallSnapshotResponse(r *v1.InstallSnapshotResponse) *raft.InstallSnapshotResponse {
	return &raft.InstallSnapshotResponse{
		RPCHeader: Header(r.Header),
		Term:      r.Term,
		Success:   r.Success,
	}
}

func InstallSnapshotResponseFrom(r *raft.InstallSnapshotResponse) *v1.InstallSnapshotResponse {
	return &v1.InstallSnapshotResponse{
		Header:  HeaderFrom(&r.RPCHeader),
		Term:    r.Term,
		Success: r.Success,
	}
}

func TimeoutNowRequest(r *v1.TimeoutNowRequest) *raft.TimeoutNowRequest {
	return &raft.TimeoutNowRequest{
		RPCHeader: Header(r.Header),
	}
}

func TimeoutNowRequestFrom(r *raft.TimeoutNowRequest) *v1.TimeoutNowRequest {
	return &v1.TimeoutNowRequest{
		Header: HeaderFrom(&r.RPCHeader),
	}
}

func TimeoutNowResponse(r *v1.TimeoutNowResponse) *raft.TimeoutNowResponse {
	return &raft.TimeoutNowResponse{
		RPCHeader: Header(r.Header),
	}
}

func TimeoutNowResponseFrom(r *raft.TimeoutNowResponse) *v1.TimeoutNowResponse {
	return &v1.TimeoutNowResponse{
		Header: HeaderFrom(&r.RPCHeader),
	}
}

func Header(r *v1.Header) raft.RPCHeader {
	return raft.RPCHeader{
		ProtocolVersion: raft.ProtocolVersion(r.Version),
		ID:              r.Id,
		Addr:            r.Addr,
	}
}

func HeaderFrom(r *raft.RPCHeader) *v1.Header {
	return &v1.Header{
		Version: v1.Header_Version(r.ProtocolVersion),
		Id:      r.ID,
		Addr:    r.Addr,
	}
}

func RaftLog(r *v1.RaftLog) *raft.Log {
	return &raft.Log{
		Index:      r.Index,
		Term:       r.Term,
		Type:       raft.LogType(r.Type),
		Data:       r.Data,
		Extensions: r.Extensions,
		AppendedAt: r.AppendedAt.AsTime(),
	}
}

func RaftLogFrom(r *raft.Log) *v1.RaftLog {
	return &v1.RaftLog{
		Index:      r.Index,
		Term:       r.Term,
		Type:       v1.RaftLog_Type(r.Type),
		Data:       r.Data,
		Extensions: r.Extensions,
		AppendedAt: timestamppb.New(r.AppendedAt),
	}
}

func RaftEntry(r *v1.RaftEntry) *badger.Entry {
	e := badger.NewEntry(r.Key, r.Value)
	if r.Expires != nil {
		e.WithTTL(r.Expires.AsDuration())
	}
	return e
}

func RaftEntryFrom(key, value []byte, ttl time.Duration) *v1.RaftEntry {
	e := &v1.RaftEntry{
		Key:   key,
		Value: value,
	}
	if ttl != 0 {
		e.Expires = durationpb.New(ttl)
	}
	return e
}
