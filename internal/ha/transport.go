package ha

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/hashicorp/raft"
	raftv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/raft/v1"
	"github.com/vinceanalytics/vince/internal/px"
	"google.golang.org/grpc"
)

type Transit struct {
	ctx         context.Context
	dialOptions []grpc.DialOption
	peers       sync.Map
	consume     chan raft.RPC
	heart       heart
	log         *slog.Logger
	timeout     time.Duration
	address     raft.ServerAddress
}

var _ raft.Transport = (*Transit)(nil)

func NewTransport(ctx context.Context, addr raft.ServerAddress, timeout time.Duration) *Transit {
	return &Transit{
		ctx:     ctx,
		address: addr,
		consume: make(chan raft.RPC),
		log:     slog.Default().With("component", "raft-transport"),
		timeout: timeout,
	}
}

type heart struct {
	mu sync.Mutex
	h  func(raft.RPC)
}

func (h *heart) beat(r raft.RPC) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.h != nil {
		h.h(r)
		return true
	}
	return false
}

func (t *Transit) SetHeartbeatHandler(cb func(rpc raft.RPC)) {
	t.heart.mu.Lock()
	t.heart.h = cb
	t.heart.mu.Unlock()
}

func (t *Transit) Register(s *grpc.Server) {
	raftv1.RegisterTransportServer(s, &ProtoTransit{
		tr: t,
	})
}
func (t *Transit) LocalAddr() raft.ServerAddress {
	return t.address
}

func (*Transit) EncodePeer(id raft.ServerID, p raft.ServerAddress) []byte {
	return []byte(p)
}

func (t *Transit) InstallSnapshot(id raft.ServerID, target raft.ServerAddress, args *raft.InstallSnapshotRequest, resp *raft.InstallSnapshotResponse, data io.Reader) error {
	c, err := t.peer(id, target)
	if err != nil {
		return err
	}
	ctx, cancel := t.context()
	defer cancel()
	stream, err := c.InstallSnapshot(ctx)
	if err != nil {
		return err
	}
	if err := stream.Send(px.InstallSnapshotRequestFrom(args)); err != nil {
		return err
	}
	var buf [4 << 10]byte
	for {
		n, err := data.Read(buf[:])
		if err == io.EOF || (err == nil && n == 0) {
			break
		}
		if err != nil {
			return err
		}
		if err := stream.Send(&raftv1.InstallSnapshotRequest{
			Data: buf[:n],
		}); err != nil {
			return err
		}
	}
	r, err := stream.CloseAndRecv()
	if err != nil {
		return err
	}
	*resp = *px.InstallSnapshotResponse(r)
	return nil
}

func (*Transit) DecodePeer(buf []byte) raft.ServerAddress {
	return raft.ServerAddress(buf)
}

func (t *Transit) Consumer() <-chan raft.RPC {
	return t.consume
}

func (t *Transit) AppendEntriesPipeline(id raft.ServerID, target raft.ServerAddress) (raft.AppendPipeline, error) {
	return nil, raft.ErrPipelineReplicationNotSupported
}

func (t *Transit) AppendEntries(
	id raft.ServerID,
	target raft.ServerAddress,
	args *raft.AppendEntriesRequest, resp *raft.AppendEntriesResponse) error {
	c, err := t.peer(id, target)
	if err != nil {
		return err
	}
	ctx, cancel := t.context()
	defer cancel()
	r, err := c.AppendEntries(ctx, px.AppendEntriesRequestFrom(args))
	if err != nil {
		return err
	}
	*resp = *px.AppendEntriesResponse(r)
	return nil
}

func (t *Transit) RequestVote(id raft.ServerID,
	target raft.ServerAddress,
	args *raft.RequestVoteRequest, resp *raft.RequestVoteResponse) error {
	c, err := t.peer(id, target)
	if err != nil {
		return err
	}
	ctx, cancel := t.context()
	defer cancel()
	r, err := c.RequestVote(ctx, px.RequestVoteRequestFrom(args))
	if err != nil {
		return err
	}
	*resp = *px.RequestVoteResponse(r)
	return nil
}

func (t *Transit) TimeoutNow(id raft.ServerID, target raft.ServerAddress, args *raft.TimeoutNowRequest, resp *raft.TimeoutNowResponse) error {
	c, err := t.peer(id, target)
	if err != nil {
		return err
	}
	ctx, cancel := t.context()
	defer cancel()
	r, err := c.TimeoutNow(ctx, px.TimeoutNowRequestFrom(args))
	if err != nil {
		return err
	}
	*resp = *px.TimeoutNowResponse(r)
	return nil
}

type RPCRequest interface {
	*raft.AppendEntriesRequest |
		*raft.RequestVoteRequest |
		*raft.InstallSnapshotRequest |
		*raft.TimeoutNowRequest
}

type RPCResponse interface {
	*raft.AppendEntriesResponse |
		*raft.RequestVoteResponse |
		*raft.InstallSnapshotResponse |
		*raft.TimeoutNowResponse
}

func (t *Transit) context() (context.Context, context.CancelFunc) {
	if t.timeout == 0 {
		return t.ctx, func() {}
	}
	return context.WithTimeout(t.ctx, t.timeout)
}

func (t *Transit) Close() error {
	var err []error
	t.peers.Range(func(key, value any) bool {
		t.peers.Delete(key)
		x := value.(*conn)
		err = append(err, x.clientConn.Close())
		return true
	})
	return errors.Join(err...)
}

func (t *Transit) peer(id raft.ServerID,
	target raft.ServerAddress) (raftv1.TransportClient, error) {
	if c, ok := t.peers.Load(id); ok {
		x := c.(*conn)
		return x.client, nil
	}
	n, err := grpc.Dial(string(target), t.dialOptions...)
	if err != nil {
		return nil, err
	}
	x := &conn{clientConn: n, client: raftv1.NewTransportClient(n)}
	t.peers.Store(id, x)
	return x.client, nil
}

type conn struct {
	clientConn *grpc.ClientConn
	client     raftv1.TransportClient
}

type ProtoTransit struct {
	tr *Transit
	raftv1.UnsafeTransportServer
}

var _ raftv1.TransportServer = (*ProtoTransit)(nil)

func (t *ProtoTransit) AppendEntries(ctx context.Context, req *raftv1.AppendEntriesRequest) (*raftv1.AppendEntriesResponse, error) {
	resp, err := t.handle(px.AppendEntriesRequest(req), nil)
	if err != nil {
		return nil, err
	}
	return px.AppendEntriesResponseFrom(resp.(*raft.AppendEntriesResponse)), nil
}
func (t *ProtoTransit) RequestVote(ctx context.Context, req *raftv1.RequestVoteRequest) (*raftv1.RequestVoteResponse, error) {
	resp, err := t.handle(px.RequestVoteRequest(req), nil)
	if err != nil {
		return nil, err
	}
	return px.RequestVoteResponseFrom(resp.(*raft.RequestVoteResponse)), nil
}

func (t *ProtoTransit) TimeoutNow(ctx context.Context, req *raftv1.TimeoutNowRequest) (*raftv1.TimeoutNowResponse, error) {
	resp, err := t.handle(px.TimeoutNowRequest(req), nil)
	if err != nil {
		return nil, err
	}
	return px.TimeoutNowResponseFrom(resp.(*raft.TimeoutNowResponse)), nil
}

func (t *ProtoTransit) InstallSnapshot(s raftv1.Transport_InstallSnapshotServer) error {
	isr, err := s.Recv()
	if err != nil {
		return err
	}
	resp, err := t.handle(px.InstallSnapshotRequest(isr), &snapshotStream{s, isr.GetData()})
	if err != nil {
		return err
	}
	return s.SendAndClose(px.InstallSnapshotResponseFrom(resp.(*raft.InstallSnapshotResponse)))
}

type snapshotStream struct {
	s raftv1.Transport_InstallSnapshotServer

	buf []byte
}

func (s *snapshotStream) Read(b []byte) (int, error) {
	if len(s.buf) > 0 {
		n := copy(b, s.buf)
		s.buf = s.buf[n:]
		return n, nil
	}
	m, err := s.s.Recv()
	if err != nil {
		return 0, err
	}
	n := copy(b, m.GetData())
	if n < len(m.GetData()) {
		s.buf = m.GetData()[n:]
	}
	return n, nil
}

func (t *ProtoTransit) AppendEntriesPipeline(s raftv1.Transport_AppendEntriesPipelineServer) error {
	for {
		msg, err := s.Recv()
		if err != nil {
			return err
		}
		resp, err := t.handle(px.AppendEntriesRequest(msg), nil)
		if err != nil {
			return err
		}
		if err := s.Send(px.AppendEntriesResponseFrom(resp.(*raft.AppendEntriesResponse))); err != nil {
			return err
		}
	}
}

func (t *ProtoTransit) handle(command interface{}, data io.Reader) (interface{}, error) {
	ch := make(chan raft.RPCResponse, 1)
	rpc := raft.RPC{
		Command:  command,
		RespChan: ch,
		Reader:   data,
	}
	if isHeartbeat(command) {
		if t.tr.heart.beat(rpc) {
			goto wait
		}
	}
	t.tr.consume <- rpc
wait:
	resp := <-ch
	if resp.Error != nil {
		return nil, resp.Error
	}
	return resp.Response, nil
}

func isHeartbeat(command interface{}) bool {
	req, ok := command.(*raft.AppendEntriesRequest)
	if !ok {
		return false
	}
	return req.Term != 0 && len(req.Leader) != 0 && req.PrevLogEntry == 0 && req.PrevLogTerm == 0 && len(req.Entries) == 0 && req.LeaderCommitIndex == 0
}
