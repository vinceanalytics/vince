package v1

import (
	"github.com/hashicorp/raft"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (r *Raft_RPC_Command) To() any {
	switch e := r.Kind.(type) {
	case *Raft_RPC_Command_AppendEntries_:
		switch o := e.AppendEntries.Kind.(type) {
		case *Raft_RPC_Command_AppendEntries_Request_:
			return o.Request.To()
		case *Raft_RPC_Command_AppendEntries_Response_:
			return o.Response.To()
		default:
			panic("unreachable")
		}
	case *Raft_RPC_Command_Vote_:
		switch o := e.Vote.Kind.(type) {
		case *Raft_RPC_Command_Vote_Request_:
			return o.Request.To()
		case *Raft_RPC_Command_Vote_Response_:
			return o.Response.To()
		default:
			panic("unreachable")
		}
	case *Raft_RPC_Command_InstallSnapshot_:
		switch o := e.InstallSnapshot.Kind.(type) {
		case *Raft_RPC_Command_InstallSnapshot_Request_:
			return o.Request.To()
		case *Raft_RPC_Command_InstallSnapshot_Response_:
			return o.Response.To()
		default:
			panic("unreachable")
		}
	case *Raft_RPC_Command_TimeoutNow_:
		switch o := e.TimeoutNow.Kind.(type) {
		case *Raft_RPC_Command_TimeoutNow_Request_:
			return o.Request.To()
		case *Raft_RPC_Command_TimeoutNow_Response_:
			return o.Response.To()
		default:
			panic("unreachable")
		}
	default:
		panic("unreachable")
	}
}

func Raft_RPC_CommandFrom(a any) *Raft_RPC_Command {
	switch e := a.(type) {
	case *raft.AppendEntriesRequest:
		return &Raft_RPC_Command{
			Kind: &Raft_RPC_Command_AppendEntries_{
				AppendEntries: &Raft_RPC_Command_AppendEntries{
					Kind: &Raft_RPC_Command_AppendEntries_Request_{
						Request: Raft_RPC_Command_AppendEntries_RequestFrom(e),
					},
				},
			},
		}
	case *raft.AppendEntriesResponse:
		return &Raft_RPC_Command{
			Kind: &Raft_RPC_Command_AppendEntries_{
				AppendEntries: &Raft_RPC_Command_AppendEntries{
					Kind: &Raft_RPC_Command_AppendEntries_Response_{
						Response: Raft_RPC_Command_AppendEntries_ResponseFrom(e),
					},
				},
			},
		}

	case *raft.RequestVoteRequest:
		return &Raft_RPC_Command{
			Kind: &Raft_RPC_Command_Vote_{
				Vote: &Raft_RPC_Command_Vote{
					Kind: &Raft_RPC_Command_Vote_Request_{
						Request: Raft_RPC_Command_Vote_RequestFrom(e),
					},
				},
			},
		}
	case *raft.RequestVoteResponse:
		return &Raft_RPC_Command{
			Kind: &Raft_RPC_Command_Vote_{
				Vote: &Raft_RPC_Command_Vote{
					Kind: &Raft_RPC_Command_Vote_Response_{
						Response: Raft_RPC_Command_Vote_ResponseFrom(e),
					},
				},
			},
		}
	case *raft.InstallSnapshotRequest:
		return &Raft_RPC_Command{
			Kind: &Raft_RPC_Command_InstallSnapshot_{
				InstallSnapshot: &Raft_RPC_Command_InstallSnapshot{
					Kind: &Raft_RPC_Command_InstallSnapshot_Request_{
						Request: Raft_RPC_Command_InstallSnapshot_RequestFrom(e),
					},
				},
			},
		}
	case *raft.InstallSnapshotResponse:
		return &Raft_RPC_Command{
			Kind: &Raft_RPC_Command_InstallSnapshot_{
				InstallSnapshot: &Raft_RPC_Command_InstallSnapshot{
					Kind: &Raft_RPC_Command_InstallSnapshot_Response_{
						Response: Raft_RPC_Command_InstallSnapshot_ResponseFrom(e),
					},
				},
			},
		}
	case *raft.TimeoutNowRequest:
		return &Raft_RPC_Command{
			Kind: &Raft_RPC_Command_TimeoutNow_{
				TimeoutNow: &Raft_RPC_Command_TimeoutNow{
					Kind: &Raft_RPC_Command_TimeoutNow_Request_{
						Request: Raft_RPC_Command_TimeoutNow_RequestFrom(e),
					},
				},
			},
		}
	case *raft.TimeoutNowResponse:
		return &Raft_RPC_Command{
			Kind: &Raft_RPC_Command_TimeoutNow_{
				TimeoutNow: &Raft_RPC_Command_TimeoutNow{
					Kind: &Raft_RPC_Command_TimeoutNow_Response_{
						Response: Raft_RPC_Command_TimeoutNow_ResponseFrom(e),
					},
				},
			},
		}
	default:
		return nil
	}
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
