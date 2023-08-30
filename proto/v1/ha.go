package v1

import (
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/hashicorp/raft"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (r *Raft_RPC_Call_Request) To() any {
	switch e := r.Kind.(type) {
	case *Raft_RPC_Call_Request_AppendEntries:
		return e.AppendEntries.To()
	case *Raft_RPC_Call_Request_Vote:
		return e.Vote.To()
	case *Raft_RPC_Call_Request_InstallSnapshot:
		return e.InstallSnapshot.To()
	case *Raft_RPC_Call_Request_TimeoutNow:
		return e.TimeoutNow.To()
	default:
		panic("unreachable")
	}
}

func Raft_RPC_Call_RequestFrom(a any) *Raft_RPC_Call_Request {
	switch e := a.(type) {
	case *raft.AppendEntriesRequest:
		return &Raft_RPC_Call_Request{
			Kind: &Raft_RPC_Call_Request_AppendEntries{
				AppendEntries: Raft_RPC_Command_AppendEntries_RequestFrom(e),
			},
		}
	case *raft.RequestVoteRequest:
		return &Raft_RPC_Call_Request{
			Kind: &Raft_RPC_Call_Request_Vote{
				Vote: Raft_RPC_Command_Vote_RequestFrom(e),
			},
		}
	case *raft.InstallSnapshotRequest:
		return &Raft_RPC_Call_Request{
			Kind: &Raft_RPC_Call_Request_InstallSnapshot{
				InstallSnapshot: Raft_RPC_Command_InstallSnapshot_RequestFrom(e),
			},
		}
	case *raft.TimeoutNowRequest:
		return &Raft_RPC_Call_Request{
			Kind: &Raft_RPC_Call_Request_TimeoutNow{
				TimeoutNow: Raft_RPC_Command_TimeoutNow_RequestFrom(e),
			},
		}
	default:
		panic("unreachable")
	}
}

func (r *Raft_RPC_Call_Response) To() any {
	switch e := r.Kind.(type) {
	case *Raft_RPC_Call_Response_AppendEntries:
		return e.AppendEntries.To()
	case *Raft_RPC_Call_Response_Vote:
		return e.Vote.To()
	case *Raft_RPC_Call_Response_InstallSnapshot:
		return e.InstallSnapshot.To()
	case *Raft_RPC_Call_Response_TimeoutNow:
		return e.TimeoutNow.To()
	default:
		panic("unreachable")
	}
}

func Raft_RPC_Call_ResponseFrom(a raft.RPCResponse) (o *Raft_RPC_Call_Response) {
	o = &Raft_RPC_Call_Response{}
	if a.Error != nil {
		o.Error = a.Error.Error()
	}
	switch e := a.Response.(type) {
	case *raft.AppendEntriesResponse:
		o.Kind = &Raft_RPC_Call_Response_AppendEntries{
			AppendEntries: Raft_RPC_Command_AppendEntries_ResponseFrom(e),
		}
	case *raft.RequestVoteResponse:
		o.Kind = &Raft_RPC_Call_Response_Vote{
			Vote: Raft_RPC_Command_Vote_ResponseFrom(e),
		}
	case *raft.InstallSnapshotResponse:
		o.Kind = &Raft_RPC_Call_Response_InstallSnapshot{
			InstallSnapshot: Raft_RPC_Command_InstallSnapshot_ResponseFrom(e),
		}
	case *raft.TimeoutNowResponse:
		o.Kind = &Raft_RPC_Call_Response_TimeoutNow{
			TimeoutNow: Raft_RPC_Command_TimeoutNow_ResponseFrom(e),
		}
	}
	return
}

func (r *Raft_RPC_Command_AppendEntries_Request) To() *raft.AppendEntriesRequest {
	entries := make([]*raft.Log, len(r.Entries))
	for i := range entries {
		entries[i] = r.Entries[i].To()
	}
	return &raft.AppendEntriesRequest{
		RPCHeader:         r.Header.To(),
		Term:              r.Term,
		PrevLogEntry:      r.PrevLogEntry,
		PrevLogTerm:       r.PrevLogTerm,
		Entries:           entries,
		LeaderCommitIndex: r.LeaderCommitIndex,
	}
}

func Raft_RPC_Command_AppendEntries_RequestFrom(r *raft.AppendEntriesRequest) *Raft_RPC_Command_AppendEntries_Request {
	entries := make([]*Raft_Log, len(r.Entries))
	for i := range entries {
		entries[i] = Raft_LogFrom(r.Entries[i])
	}
	return &Raft_RPC_Command_AppendEntries_Request{
		Header:            Raft_RPC_Command_HeaderFrom(&r.RPCHeader),
		Term:              r.Term,
		PrevLogEntry:      r.PrevLogEntry,
		PrevLogTerm:       r.PrevLogTerm,
		Entries:           entries,
		LeaderCommitIndex: r.LeaderCommitIndex,
	}
}
func (r *Raft_RPC_Command_AppendEntries_Response) To() *raft.AppendEntriesResponse {
	return &raft.AppendEntriesResponse{
		RPCHeader:      r.Header.To(),
		Term:           r.Term,
		LastLog:        r.LastLog,
		Success:        r.Success,
		NoRetryBackoff: r.NoRetryBackoff,
	}
}

func Raft_RPC_Command_AppendEntries_ResponseFrom(r *raft.AppendEntriesResponse) *Raft_RPC_Command_AppendEntries_Response {
	return &Raft_RPC_Command_AppendEntries_Response{
		Header:         Raft_RPC_Command_HeaderFrom(&r.RPCHeader),
		Term:           r.Term,
		LastLog:        r.LastLog,
		Success:        r.Success,
		NoRetryBackoff: r.NoRetryBackoff,
	}
}

func (r *Raft_RPC_Command_Vote_Request) To() *raft.RequestVoteRequest {
	return &raft.RequestVoteRequest{
		RPCHeader:          r.Header.To(),
		Term:               r.Term,
		LastLogIndex:       r.LastLogIndex,
		LastLogTerm:        r.LastLogTerm,
		LeadershipTransfer: r.LeadershipTransfer,
	}
}

func Raft_RPC_Command_Vote_RequestFrom(r *raft.RequestVoteRequest) *Raft_RPC_Command_Vote_Request {
	return &Raft_RPC_Command_Vote_Request{
		Header:             Raft_RPC_Command_HeaderFrom(&r.RPCHeader),
		Term:               r.Term,
		LastLogIndex:       r.LastLogIndex,
		LastLogTerm:        r.LastLogTerm,
		LeadershipTransfer: r.LeadershipTransfer,
	}
}

func (r *Raft_RPC_Command_Vote_Response) To() *raft.RequestVoteResponse {
	return &raft.RequestVoteResponse{
		RPCHeader: r.Header.To(),
		Term:      r.Term,
		Granted:   r.Granted,
	}
}

func Raft_RPC_Command_Vote_ResponseFrom(r *raft.RequestVoteResponse) *Raft_RPC_Command_Vote_Response {
	return &Raft_RPC_Command_Vote_Response{
		Header:  Raft_RPC_Command_HeaderFrom(&r.RPCHeader),
		Term:    r.Term,
		Granted: r.Granted,
	}
}

func (r *Raft_RPC_Command_InstallSnapshot_Request) To() *raft.InstallSnapshotRequest {
	return &raft.InstallSnapshotRequest{
		RPCHeader:          r.Header.To(),
		Term:               r.Term,
		Leader:             r.Leader,
		LastLogIndex:       r.LastLogIndex,
		LastLogTerm:        r.LastLogTerm,
		Configuration:      r.Configuration,
		ConfigurationIndex: r.ConfigurationIndex,
		Size:               r.Size,
	}
}

func Raft_RPC_Command_InstallSnapshot_RequestFrom(r *raft.InstallSnapshotRequest) *Raft_RPC_Command_InstallSnapshot_Request {
	return &Raft_RPC_Command_InstallSnapshot_Request{
		Header:             Raft_RPC_Command_HeaderFrom(&r.RPCHeader),
		Term:               r.Term,
		Leader:             r.Leader,
		LastLogIndex:       r.LastLogIndex,
		LastLogTerm:        r.LastLogTerm,
		Configuration:      r.Configuration,
		ConfigurationIndex: r.ConfigurationIndex,
		Size:               r.Size,
	}
}
func (r *Raft_RPC_Command_InstallSnapshot_Response) To() *raft.InstallSnapshotResponse {
	return &raft.InstallSnapshotResponse{
		RPCHeader: r.Header.To(),
		Term:      r.Term,
		Success:   r.Success,
	}
}

func Raft_RPC_Command_InstallSnapshot_ResponseFrom(r *raft.InstallSnapshotResponse) *Raft_RPC_Command_InstallSnapshot_Response {
	return &Raft_RPC_Command_InstallSnapshot_Response{
		Header:  Raft_RPC_Command_HeaderFrom(&r.RPCHeader),
		Term:    r.Term,
		Success: r.Success,
	}
}

func (r *Raft_RPC_Command_TimeoutNow_Request) To() *raft.TimeoutNowRequest {
	return &raft.TimeoutNowRequest{
		RPCHeader: r.Header.To(),
	}
}

func Raft_RPC_Command_TimeoutNow_RequestFrom(r *raft.TimeoutNowRequest) *Raft_RPC_Command_TimeoutNow_Request {
	return &Raft_RPC_Command_TimeoutNow_Request{
		Header: Raft_RPC_Command_HeaderFrom(&r.RPCHeader),
	}
}

func (r *Raft_RPC_Command_TimeoutNow_Response) To() *raft.TimeoutNowResponse {
	return &raft.TimeoutNowResponse{
		RPCHeader: r.Header.To(),
	}
}

func Raft_RPC_Command_TimeoutNow_ResponseFrom(r *raft.TimeoutNowResponse) *Raft_RPC_Command_TimeoutNow_Response {
	return &Raft_RPC_Command_TimeoutNow_Response{
		Header: Raft_RPC_Command_HeaderFrom(&r.RPCHeader),
	}
}

func (r *Raft_RPC_Command_Header) To() raft.RPCHeader {
	return raft.RPCHeader{
		ProtocolVersion: raft.ProtocolVersion(r.Version),
		ID:              r.Id,
		Addr:            r.Addr,
	}
}

func Raft_RPC_Command_HeaderFrom(r *raft.RPCHeader) *Raft_RPC_Command_Header {
	return &Raft_RPC_Command_Header{
		Version: Raft_RPC_Command_Header_Version(r.ProtocolVersion),
		Id:      r.ID,
		Addr:    r.Addr,
	}
}

func (r *Raft_Log) To() *raft.Log {
	return &raft.Log{
		Index:      r.Index,
		Term:       r.Term,
		Type:       raft.LogType(r.Type),
		Data:       r.Data,
		Extensions: r.Extensions,
		AppendedAt: r.AppendedAt.AsTime(),
	}
}

func Raft_LogFrom(r *raft.Log) *Raft_Log {
	return &Raft_Log{
		Index:      r.Index,
		Term:       r.Term,
		Type:       Raft_Log_Type(r.Type),
		Data:       r.Data,
		Extensions: r.Extensions,
		AppendedAt: timestamppb.New(r.AppendedAt),
	}
}

func (r *Raft_Entry) To() *badger.Entry {
	e := badger.NewEntry(r.Key, r.Value)
	if r.Expires != nil {
		e.WithTTL(r.Expires.AsDuration())
	}
	return e
}

func Raft_EntryFrom(key, value []byte, ttl time.Duration) *Raft_Entry {
	e := &Raft_Entry{
		Key:   key,
		Value: value,
	}
	if ttl != 0 {
		e.Expires = durationpb.New(ttl)
	}
	return e
}
