package ha

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"

	"github.com/hashicorp/raft"
	v1 "github.com/vinceanalytics/vince/proto/v1"
	"google.golang.org/protobuf/proto"
	"nhooyr.io/websocket"
)

type Transit struct {
	ctx     context.Context
	peers   sync.Map
	consume chan raft.RPC
	id      raft.ServerID
	heart   heart
	dia     func(context.Context,
		raft.ServerID, raft.ServerAddress) (*websocket.Conn, error)
}

var _ raft.Transport = (*Transit)(nil)

type heart struct {
	mu sync.Mutex
	h  func(raft.RPC)
}

func (h *heart) beat(r raft.RPC) {
	h.mu.Lock()
	if h.h != nil {
		h.h(r)
	}
	h.mu.Unlock()
}

func (t *Transit) SetHeartbeatHandler(cb func(rpc raft.RPC)) {
	t.heart.mu.Lock()
	t.heart.h = cb
	t.heart.mu.Unlock()
}

func (t *Transit) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{})
	if err != nil {
		slog.Error("failed accepting websocket connection", "err", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	log := slog.Default().With(
		slog.String("component", "raft"),
	)
	for {
		select {
		case <-t.ctx.Done():
			log.Info("websocket raft transport is closed")
			return
		default:
			var req v1.Raft_RPC_Call_Request
			err := read(t.ctx, conn, &req)
			if err != nil {
				log.Error("failed reading rpc request", "err", err.Error())
				return
			}
			result := make(chan raft.RPCResponse, 1)

			rx := raft.RPC{
				Command:  req.To(),
				RespChan: result,
			}

			var beat bool
			if x, ok := req.Kind.(*v1.Raft_RPC_Call_Request_AppendEntries); ok {
				a := x.AppendEntries
				beat = a.Term != 0 && a.Header.Addr != nil &&
					a.PrevLogEntry == 0 && a.PrevLogTerm == 0 &&
					len(a.Entries) == 0 && a.LeaderCommitIndex == 0
			}
			if beat {
				t.heart.beat(rx)
			} else {
				select {
				case <-t.ctx.Done():
					log.Info("websocket raft transport is closed")
					return
				case t.consume <- rx:
				}
			}
			select {
			case <-t.ctx.Done():
				log.Info("websocket raft transport is closed")
				return
			case res := <-result:
				o := v1.Raft_RPC_Call_ResponseFrom(res)
				err = write(t.ctx, conn, o)
				if err != nil {
					log.Error("failed writing to websocket connection", "err", err.Error())
					return
				}
			}
		}

	}
}

func (t *Transit) LocalAddr() raft.ServerAddress {
	return raft.ServerAddress(t.id)
}

func (*Transit) EncodePeer(id raft.ServerID, p raft.ServerAddress) []byte {
	return []byte(id)
}

func (t *Transit) InstallSnapshot(id raft.ServerID, target raft.ServerAddress, args *raft.InstallSnapshotRequest, resp *raft.InstallSnapshotResponse, data io.Reader) error {
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
	return rpc(t, id, target, args, resp)
}

func (t *Transit) RequestVote(id raft.ServerID,
	target raft.ServerAddress,
	args *raft.RequestVoteRequest, resp *raft.RequestVoteResponse) error {
	return rpc(t, id, target, args, resp)
}

func (t *Transit) TimeoutNow(id raft.ServerID, target raft.ServerAddress, args *raft.TimeoutNowRequest, resp *raft.TimeoutNowResponse) error {
	return rpc(t, id, target, args, resp)
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

func rpc[Req RPCRequest, Res RPCResponse](t *Transit, id raft.ServerID,
	target raft.ServerAddress, args Req, resp Res) error {
	result := t.send(id, target, v1.Raft_RPC_Call_RequestFrom(args))
	if result.Error != "" {
		return errors.New(result.Error)
	}
	r := result.To()
	switch e := any(resp).(type) {
	case *raft.AppendEntriesResponse:
		*e = *r.(*raft.AppendEntriesResponse)
	case *raft.RequestVoteResponse:
		*e = *r.(*raft.RequestVoteResponse)
	case *raft.InstallSnapshotResponse:
		*e = *r.(*raft.InstallSnapshotResponse)
	case *raft.TimeoutNowResponse:
		*e = *r.(*raft.TimeoutNowResponse)
	}
	return nil
}

func (t *Transit) send(id raft.ServerID,
	target raft.ServerAddress, r *v1.Raft_RPC_Call_Request) *v1.Raft_RPC_Call_Response {
	conn, err := t.peer(id, target)
	if err != nil {
		return &v1.Raft_RPC_Call_Response{
			Error: err.Error(),
		}
	}
	ctx := t.context()
	err = write(ctx, conn, r)
	if err != nil {
		return &v1.Raft_RPC_Call_Response{
			Error: err.Error(),
		}
	}
	var o v1.Raft_RPC_Call_Response
	err = read(ctx, conn, &o)
	if err != nil {
		return &v1.Raft_RPC_Call_Response{
			Error: err.Error(),
		}
	}
	return &o
}

func (t *Transit) context() context.Context {
	return context.Background()
}

func (t *Transit) Close() error {
	t.peers.Range(func(key, value any) bool {
		t.peers.Delete(key)
		value.(*websocket.Conn).Close(websocket.StatusGoingAway, "closing transport")
		return true
	})
	return nil
}

func (t *Transit) peer(id raft.ServerID,
	target raft.ServerAddress) (*websocket.Conn, error) {
	if conn, ok := t.peers.Load(id); ok {
		return conn.(*websocket.Conn), nil
	}
	conn, err := t.dia(t.ctx, id, target)
	if err != nil {
		return nil, err
	}
	t.peers.Store(id, conn)
	return conn, nil
}

func read(ctx context.Context, c *websocket.Conn, m proto.Message) error {
	typ, r, err := c.Reader(ctx)
	if err != nil {
		return err
	}
	if typ != websocket.MessageBinary {
		c.Close(websocket.StatusUnsupportedData, "expected binary message")
		return fmt.Errorf("expected binary message for protobuf but got: %v", typ)
	}
	b := pool.Get().(*bytes.Buffer)
	defer func() {
		b.Reset()
		pool.Put(b)
	}()
	_, err = b.ReadFrom(r)
	if err != nil {
		return err
	}
	err = proto.Unmarshal(b.Bytes(), m)
	if err != nil {
		c.Close(websocket.StatusInvalidFramePayloadData, "failed to unmarshal protobuf")
		return fmt.Errorf("failed to unmarshal protobuf: %w", err)
	}
	return nil
}

func write(ctx context.Context, c *websocket.Conn, m proto.Message) error {
	b, err := proto.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal protobuf: %w", err)
	}
	return c.Write(ctx, websocket.MessageBinary, b)
}

var pool = &sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}
