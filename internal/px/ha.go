package px

import (
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/hashicorp/raft"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/raft/v1"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Raft_RPC_Call_RequestTo(r *v1.Raft_RPC_Call_Request) any {
	switch e := r.Kind.(type) {
	case *v1.Raft_RPC_Call_Request_AppendEntries:
		return Raft_RPC_Command_AppendEntries_RequestTo(e.AppendEntries)
	case *v1.Raft_RPC_Call_Request_Vote:
		return Raft_RPC_Command_Vote_RequestTo(e.Vote)
	case *v1.Raft_RPC_Call_Request_InstallSnapshot:
		return Raft_RPC_Command_InstallSnapshot_RequestTo(e.InstallSnapshot)
	case *v1.Raft_RPC_Call_Request_TimeoutNow:
		return Raft_RPC_Command_TimeoutNow_RequestTo(e.TimeoutNow)
	default:
		panic("unreachable")
	}
}

func Raft_RPC_Call_RequestFrom(a any) *v1.Raft_RPC_Call_Request {
	switch e := a.(type) {
	case *raft.AppendEntriesRequest:
		return &v1.Raft_RPC_Call_Request{
			Kind: &v1.Raft_RPC_Call_Request_AppendEntries{
				AppendEntries: Raft_RPC_Command_AppendEntries_RequestFrom(e),
			},
		}
	case *raft.RequestVoteRequest:
		return &v1.Raft_RPC_Call_Request{
			Kind: &v1.Raft_RPC_Call_Request_Vote{
				Vote: Raft_RPC_Command_Vote_RequestFrom(e),
			},
		}
	case *raft.InstallSnapshotRequest:
		return &v1.Raft_RPC_Call_Request{
			Kind: &v1.Raft_RPC_Call_Request_InstallSnapshot{
				InstallSnapshot: Raft_RPC_Command_InstallSnapshot_RequestFrom(e),
			},
		}
	case *raft.TimeoutNowRequest:
		return &v1.Raft_RPC_Call_Request{
			Kind: &v1.Raft_RPC_Call_Request_TimeoutNow{
				TimeoutNow: Raft_RPC_Command_TimeoutNow_RequestFrom(e),
			},
		}
	default:
		panic("unreachable")
	}
}

func Raft_RPC_Call_ResponseTo(r *v1.Raft_RPC_Call_Response) any {
	switch e := r.Kind.(type) {
	case *v1.Raft_RPC_Call_Response_AppendEntries:
		return Raft_RPC_Command_AppendEntries_ResponseTo(e.AppendEntries)
	case *v1.Raft_RPC_Call_Response_Vote:
		return Raft_RPC_Command_Vote_ResponseTo(e.Vote)
	case *v1.Raft_RPC_Call_Response_InstallSnapshot:
		return Raft_RPC_Command_InstallSnapshot_ResponseTo(e.InstallSnapshot)
	case *v1.Raft_RPC_Call_Response_TimeoutNow:
		return Raft_RPC_Command_TimeoutNow_ResponseTo(e.TimeoutNow)
	default:
		panic("unreachable")
	}
}

func Raft_RPC_Call_ResponseFrom(a raft.RPCResponse) (o *v1.Raft_RPC_Call_Response) {
	o = &v1.Raft_RPC_Call_Response{}
	if a.Error != nil {
		o.Error = a.Error.Error()
	}
	switch e := a.Response.(type) {
	case *raft.AppendEntriesResponse:
		o.Kind = &v1.Raft_RPC_Call_Response_AppendEntries{
			AppendEntries: Raft_RPC_Command_AppendEntries_ResponseFrom(e),
		}
	case *raft.RequestVoteResponse:
		o.Kind = &v1.Raft_RPC_Call_Response_Vote{
			Vote: Raft_RPC_Command_Vote_ResponseFrom(e),
		}
	case *raft.InstallSnapshotResponse:
		o.Kind = &v1.Raft_RPC_Call_Response_InstallSnapshot{
			InstallSnapshot: Raft_RPC_Command_InstallSnapshot_ResponseFrom(e),
		}
	case *raft.TimeoutNowResponse:
		o.Kind = &v1.Raft_RPC_Call_Response_TimeoutNow{
			TimeoutNow: Raft_RPC_Command_TimeoutNow_ResponseFrom(e),
		}
	}
	return
}

func Raft_RPC_Command_AppendEntries_RequestTo(r *v1.Raft_RPC_Command_AppendEntries_Request) *raft.AppendEntriesRequest {
	entries := make([]*raft.Log, len(r.Entries))
	for i := range entries {
		entries[i] = Raft_LogTo(r.Entries[i])
	}
	return &raft.AppendEntriesRequest{
		RPCHeader:         Raft_RPC_Command_HeaderTo(r.Header),
		Term:              r.Term,
		PrevLogEntry:      r.PrevLogEntry,
		PrevLogTerm:       r.PrevLogTerm,
		Entries:           entries,
		LeaderCommitIndex: r.LeaderCommitIndex,
	}
}

func Raft_RPC_Command_AppendEntries_RequestFrom(r *raft.AppendEntriesRequest) *v1.Raft_RPC_Command_AppendEntries_Request {
	entries := make([]*v1.Raft_Log, len(r.Entries))
	for i := range entries {
		entries[i] = Raft_LogFrom(r.Entries[i])
	}
	return &v1.Raft_RPC_Command_AppendEntries_Request{
		Header:            Raft_RPC_Command_HeaderFrom(&r.RPCHeader),
		Term:              r.Term,
		PrevLogEntry:      r.PrevLogEntry,
		PrevLogTerm:       r.PrevLogTerm,
		Entries:           entries,
		LeaderCommitIndex: r.LeaderCommitIndex,
	}
}

func Raft_RPC_Command_AppendEntries_ResponseTo(r *v1.Raft_RPC_Command_AppendEntries_Response) *raft.AppendEntriesResponse {
	return &raft.AppendEntriesResponse{
		RPCHeader:      Raft_RPC_Command_HeaderTo(r.Header),
		Term:           r.Term,
		LastLog:        r.LastLog,
		Success:        r.Success,
		NoRetryBackoff: r.NoRetryBackoff,
	}
}

func Raft_RPC_Command_AppendEntries_ResponseFrom(r *raft.AppendEntriesResponse) *v1.Raft_RPC_Command_AppendEntries_Response {
	return &v1.Raft_RPC_Command_AppendEntries_Response{
		Header:         Raft_RPC_Command_HeaderFrom(&r.RPCHeader),
		Term:           r.Term,
		LastLog:        r.LastLog,
		Success:        r.Success,
		NoRetryBackoff: r.NoRetryBackoff,
	}
}

func Raft_RPC_Command_Vote_RequestTo(r *v1.Raft_RPC_Command_Vote_Request) *raft.RequestVoteRequest {
	return &raft.RequestVoteRequest{
		RPCHeader:          Raft_RPC_Command_HeaderTo(r.Header),
		Term:               r.Term,
		LastLogIndex:       r.LastLogIndex,
		LastLogTerm:        r.LastLogTerm,
		LeadershipTransfer: r.LeadershipTransfer,
	}
}

func Raft_RPC_Command_Vote_RequestFrom(r *raft.RequestVoteRequest) *v1.Raft_RPC_Command_Vote_Request {
	return &v1.Raft_RPC_Command_Vote_Request{
		Header:             Raft_RPC_Command_HeaderFrom(&r.RPCHeader),
		Term:               r.Term,
		LastLogIndex:       r.LastLogIndex,
		LastLogTerm:        r.LastLogTerm,
		LeadershipTransfer: r.LeadershipTransfer,
	}
}

func Raft_RPC_Command_Vote_ResponseTo(r *v1.Raft_RPC_Command_Vote_Response) *raft.RequestVoteResponse {
	return &raft.RequestVoteResponse{
		RPCHeader: Raft_RPC_Command_HeaderTo(r.Header),
		Term:      r.Term,
		Granted:   r.Granted,
	}
}

func Raft_RPC_Command_Vote_ResponseFrom(r *raft.RequestVoteResponse) *v1.Raft_RPC_Command_Vote_Response {
	return &v1.Raft_RPC_Command_Vote_Response{
		Header:  Raft_RPC_Command_HeaderFrom(&r.RPCHeader),
		Term:    r.Term,
		Granted: r.Granted,
	}
}

func Raft_RPC_Command_InstallSnapshot_RequestTo(r *v1.Raft_RPC_Command_InstallSnapshot_Request) *raft.InstallSnapshotRequest {
	return &raft.InstallSnapshotRequest{
		RPCHeader:          Raft_RPC_Command_HeaderTo(r.Header),
		Term:               r.Term,
		Leader:             r.Leader,
		LastLogIndex:       r.LastLogIndex,
		LastLogTerm:        r.LastLogTerm,
		Configuration:      r.Configuration,
		ConfigurationIndex: r.ConfigurationIndex,
		Size:               r.Size,
	}
}

func Raft_RPC_Command_InstallSnapshot_RequestFrom(r *raft.InstallSnapshotRequest) *v1.Raft_RPC_Command_InstallSnapshot_Request {
	return &v1.Raft_RPC_Command_InstallSnapshot_Request{
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

func Raft_RPC_Command_InstallSnapshot_ResponseTo(r *v1.Raft_RPC_Command_InstallSnapshot_Response) *raft.InstallSnapshotResponse {
	return &raft.InstallSnapshotResponse{
		RPCHeader: Raft_RPC_Command_HeaderTo(r.Header),
		Term:      r.Term,
		Success:   r.Success,
	}
}

func Raft_RPC_Command_InstallSnapshot_ResponseFrom(r *raft.InstallSnapshotResponse) *v1.Raft_RPC_Command_InstallSnapshot_Response {
	return &v1.Raft_RPC_Command_InstallSnapshot_Response{
		Header:  Raft_RPC_Command_HeaderFrom(&r.RPCHeader),
		Term:    r.Term,
		Success: r.Success,
	}
}

func Raft_RPC_Command_TimeoutNow_RequestTo(r *v1.Raft_RPC_Command_TimeoutNow_Request) *raft.TimeoutNowRequest {
	return &raft.TimeoutNowRequest{
		RPCHeader: Raft_RPC_Command_HeaderTo(r.Header),
	}
}

func Raft_RPC_Command_TimeoutNow_RequestFrom(r *raft.TimeoutNowRequest) *v1.Raft_RPC_Command_TimeoutNow_Request {
	return &v1.Raft_RPC_Command_TimeoutNow_Request{
		Header: Raft_RPC_Command_HeaderFrom(&r.RPCHeader),
	}
}

func Raft_RPC_Command_TimeoutNow_ResponseTo(r *v1.Raft_RPC_Command_TimeoutNow_Response) *raft.TimeoutNowResponse {
	return &raft.TimeoutNowResponse{
		RPCHeader: Raft_RPC_Command_HeaderTo(r.Header),
	}
}

func Raft_RPC_Command_TimeoutNow_ResponseFrom(r *raft.TimeoutNowResponse) *v1.Raft_RPC_Command_TimeoutNow_Response {
	return &v1.Raft_RPC_Command_TimeoutNow_Response{
		Header: Raft_RPC_Command_HeaderFrom(&r.RPCHeader),
	}
}

func Raft_RPC_Command_HeaderTo(r *v1.Raft_RPC_Command_Header) raft.RPCHeader {
	return raft.RPCHeader{
		ProtocolVersion: raft.ProtocolVersion(r.Version),
		ID:              r.Id,
		Addr:            r.Addr,
	}
}

func Raft_RPC_Command_HeaderFrom(r *raft.RPCHeader) *v1.Raft_RPC_Command_Header {
	return &v1.Raft_RPC_Command_Header{
		Version: v1.Raft_RPC_Command_Header_Version(r.ProtocolVersion),
		Id:      r.ID,
		Addr:    r.Addr,
	}
}

func Raft_LogTo(r *v1.Raft_Log) *raft.Log {
	return &raft.Log{
		Index:      r.Index,
		Term:       r.Term,
		Type:       raft.LogType(r.Type),
		Data:       r.Data,
		Extensions: r.Extensions,
		AppendedAt: r.AppendedAt.AsTime(),
	}
}

func Raft_LogFrom(r *raft.Log) *v1.Raft_Log {
	return &v1.Raft_Log{
		Index:      r.Index,
		Term:       r.Term,
		Type:       v1.Raft_Log_Type(r.Type),
		Data:       r.Data,
		Extensions: r.Extensions,
		AppendedAt: timestamppb.New(r.AppendedAt),
	}
}

func Raft_EntryTo(r *v1.Raft_Entry) *badger.Entry {
	e := badger.NewEntry(r.Key, r.Value)
	if r.Expires != nil {
		e.WithTTL(r.Expires.AsDuration())
	}
	return e
}

func Raft_EntryFrom(key, value []byte, ttl time.Duration) *v1.Raft_Entry {
	e := &v1.Raft_Entry{
		Key:   key,
		Value: value,
	}
	if ttl != 0 {
		e.Expires = durationpb.New(ttl)
	}
	return e
}
