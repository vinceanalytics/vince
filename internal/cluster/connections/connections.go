package connections

import (
	"sync"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"google.golang.org/grpc"
)

type Manager struct {
	localAddr   string
	dialOptions []grpc.DialOption
	conns       sync.Map
}

func New(localAddr string, opts ...grpc.DialOption) *Manager {
	return &Manager{
		localAddr:   localAddr,
		dialOptions: opts,
	}
}

func (m *Manager) LocalAddress() string {
	return m.localAddr
}

func (m *Manager) Get(peer, target string) (*Conn, error) {
	o, ok := m.conns.Load(peer)
	if ok {
		return o.(*Conn), nil
	}
	x, err := grpc.Dial(target, m.dialOptions...)
	if err != nil {
		return nil, err
	}
	conn := &Conn{
		ClientConn:            x,
		RaftTransportClient:   v1.NewRaftTransportClient(x),
		InternalCLusterClient: v1.NewInternalCLusterClient(x),
	}
	m.conns.Store(peer, conn)
	return conn, nil
}

func (m *Manager) Close() (err error) {
	m.conns.Range(func(key, value any) bool {
		e := value.(*Conn).Close()
		if e != nil {
			err = e
		}
		return true
	})
	return
}

type Conn struct {
	*grpc.ClientConn
	v1.RaftTransportClient
	v1.InternalCLusterClient
}
