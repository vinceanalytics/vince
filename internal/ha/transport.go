package ha

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/hashicorp/raft"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/v1"
	"github.com/vinceanalytics/vince/internal/px"
	"google.golang.org/protobuf/proto"
	"nhooyr.io/websocket"
)

type Stream interface {
	Update(*v1.Raft_Config)
	Dial(context.Context,
		raft.ServerID, raft.ServerAddress) (io.ReadWriteCloser, error)
}

type Transit struct {
	ctx     context.Context
	peers   sync.Map
	consume chan raft.RPC
	svr     *v1.Raft_Config_Server
	heart   heart
	stream  Stream
	log     *slog.Logger
	timeout time.Duration
}

var _ raft.Transport = (*Transit)(nil)

func NewTransport(ctx context.Context, svr *v1.Raft_Config_Server, stream Stream, timeout time.Duration) *Transit {
	return &Transit{
		ctx:     ctx,
		consume: make(chan raft.RPC),
		svr:     svr,
		stream:  stream,
		log:     slog.Default().With("component", "raft-transport"),
		timeout: timeout,
	}
}

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
		t.log.Error("failed accepting websocket connection", "err", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	log := t.log
	log.Info("accepted raft websocket connection")
	wrap := newWrap(t.ctx, conn)
	for {
		select {
		case <-t.ctx.Done():
			log.Info("websocket raft transport is closed")
			return
		default:
			var req v1.Raft_RPC_Call_Request
			ctx, cancel := t.context()
			err := read(ctx, wrap, &req)
			cancel()
			if err != nil {
				log.Error("failed reading rpc request", "err", err.Error())
				return
			}

			result := make(chan raft.RPCResponse, 1)

			rx := raft.RPC{
				Command:  px.Raft_RPC_Call_RequestTo(&req),
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
				o := px.Raft_RPC_Call_ResponseFrom(res)
				ctx, cancel := t.context()
				err = write(ctx, wrap, o)
				cancel()
				if err != nil {
					log.Error("failed writing to websocket connection", "err", err.Error())
					return
				}
			}
		}

	}
}

func (t *Transit) LocalAddr() raft.ServerAddress {
	return raft.ServerAddress(t.svr.Address)
}

func (*Transit) EncodePeer(id raft.ServerID, p raft.ServerAddress) []byte {
	return []byte(p)
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
	result := t.send(id, target, px.Raft_RPC_Call_RequestFrom(args))
	if result.Error != "" {
		return errors.New(result.Error)
	}
	r := px.Raft_RPC_Call_ResponseTo(result)
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
	ctx, cancel := t.context()
	err = write(ctx, conn, r)
	cancel()
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

func (t *Transit) context() (context.Context, context.CancelFunc) {
	if t.timeout == 0 {
		return t.ctx, func() {}
	}
	return context.WithTimeout(t.ctx, t.timeout)
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
	target raft.ServerAddress) (io.ReadWriteCloser, error) {
	if conn, ok := t.peers.Load(id); ok {
		return conn.(io.ReadWriteCloser), nil
	}
	ctx, cancel := t.context()
	defer cancel()
	conn, err := t.stream.Dial(ctx, id, target)
	if err != nil {
		return nil, err
	}
	t.peers.Store(id, conn)
	return conn, nil
}

func read(ctx context.Context, r io.Reader, m proto.Message) error {
	b := pool.Get().(*bytes.Buffer)
	defer func() {
		b.Reset()
		pool.Put(b)
	}()
	_, err := b.ReadFrom(r)
	if err != nil {
		return err
	}
	err = proto.Unmarshal(b.Bytes(), m)
	if err != nil {
		return fmt.Errorf("failed to unmarshal protobuf: %w", err)
	}
	return nil
}

func write(ctx context.Context, c io.Writer, m proto.Message) error {
	b, err := proto.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal protobuf: %w", err)
	}
	_, err = c.Write(b)
	return err
}

var pool = &sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

type wrapSocket struct {
	ctx   context.Context
	conn  *websocket.Conn
	reset bool
	r     *bufio.Reader
}

func newWrap(ctx context.Context, c *websocket.Conn) *wrapSocket {
	return &wrapSocket{
		ctx:   ctx,
		conn:  c,
		reset: true,
		r:     bufio.NewReader(nil),
	}
}

var _ io.ReadWriteCloser = (*wrapSocket)(nil)

func (w *wrapSocket) Read(p []byte) (n int, err error) {
	if w.reset {
		typ, r, err := w.conn.Reader(w.ctx)
		if err != nil {
			return 0, err
		}
		if typ != websocket.MessageBinary {
			w.conn.Close(websocket.StatusUnsupportedData, "expected binary message")
			return 0, fmt.Errorf("expected binary message for protobuf but got: %v", typ)
		}
		w.r.Reset(r)
	}
	n, err = w.r.Read(p)
	if err != nil {
		if errors.Is(err, io.EOF) {
			w.reset = true
		}
	}
	return
}

func (w *wrapSocket) Write(p []byte) (int, error) {
	err := w.conn.Write(w.ctx, websocket.MessageBinary, p)
	return len(p), err
}

func (w *wrapSocket) Close() error {
	return w.conn.Close(websocket.StatusGoingAway, "closing transport")
}

type wsStream struct {
	peers sync.Map
}

var _ Stream = (*wsStream)(nil)

func NewWsStream() Stream {
	return &wsStream{}
}

var client = &http.Client{}

func (ws *wsStream) Update(x *v1.Raft_Config) {
	for _, svr := range x.Servers {
		ws.peers.Store(svr.Id, svr)
	}
}

func (ws *wsStream) Dial(
	ctx context.Context,
	id raft.ServerID, addr raft.ServerAddress) (io.ReadWriteCloser, error) {
	x, ok := ws.peers.Load(string(id))
	if !ok {
		return nil, fmt.Errorf("peer %s:%s can not be reached", id, addr)
	}
	s := x.(*v1.Raft_Config_Server)
	h := make(http.Header)
	h.Set("authorization", "Bearer "+s.Token)
	conn, _, err := websocket.Dial(ctx, s.Address+"/raft", &websocket.DialOptions{
		HTTPClient: client,
		HTTPHeader: h,
	})
	if err != nil {
		return nil, err
	}
	return newWrap(ctx, conn), nil
}
