package cluster

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"sync"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/cluster/auth"
	"github.com/vinceanalytics/vince/internal/cluster/http"
	"github.com/vinceanalytics/vince/internal/cluster/rtls"
	"github.com/vinceanalytics/vince/internal/cluster/tcp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/emptypb"
)

// CreateRaftDialer creates a dialer for connecting to other nodes' Raft service. If the cert and
// key arguments are not set, then the returned dialer will not use TLS.
func CreateRaftDialer(cert, key, caCert, serverName string, Insecure bool) (*tcp.Dialer, error) {
	var dialerTLSConfig *tls.Config
	var err error
	if cert != "" || key != "" {
		dialerTLSConfig, err = rtls.CreateClientConfig(cert, key, caCert, serverName, Insecure)
		if err != nil {
			return nil, fmt.Errorf("failed to create TLS config for Raft dialer: %s", err.Error())
		}
	}
	return tcp.NewDialer(MuxRaftHeader, dialerTLSConfig), nil
}

// CredentialsFor returns a Credentials instance for the given username, or nil if
// the given CredentialsStore is nil, or the username is not found.
func CredentialsFor(credStr *auth.CredentialsStore, username string) *v1.Credentials {
	if credStr == nil {
		return nil
	}
	pw, ok := credStr.Password(username)
	if !ok {
		return nil
	}
	return &v1.Credentials{
		Username: username,
		Password: pw,
	}
}

// Client allows communicating with a remote node.
type Client struct {
	dialOpts      []grpc.DialOption
	insecure      bool
	localNodeAddr string
	mu            sync.RWMutex
	clients       map[string]*Klient
}

var _ http.Cluster = (*Client)(nil)

type Klient struct {
	*grpc.ClientConn
	v1.InternalCLusterClient
}

// NewClient returns a client instance for talking to a remote node.
// Clients will retry certain commands if they fail, to allow for
// remote node restarts. Cluster management operations such as joining
// and removing nodes are not retried, to make it clear to the operator
// that the operation failed. In addition, higher-level code will
// usually retry these operations.
func NewClient(insecure bool, opts ...grpc.DialOption) *Client {
	return &Client{
		insecure: insecure,
		dialOpts: opts,
		clients:  make(map[string]*Klient),
	}
}

// SetLocal informs the client instance of the node address for the node
// using this client. Along with the Service instance it allows this
// client to serve requests for this node locally without the network hop.
func (c *Client) SetLocal(nodeAddr string, serv v1.InternalCLusterClient) error {
	c.mu.Lock()
	c.clients[nodeAddr] = &Klient{
		InternalCLusterClient: serv,
	}
	c.localNodeAddr = nodeAddr
	c.mu.Unlock()
	return nil
}

// GetNodeAPIAddr retrieves the API Address for the node at nodeAddr
func (c *Client) GetNodeAPIAddr(ctx context.Context, nodeAddr string) (string, error) {
	remote, err := c.node(nodeAddr)
	if err != nil {
		return "", err
	}
	meta, err := remote.NodeAPI(ctx, &emptypb.Empty{})
	if err != nil {
		return "", err
	}
	return meta.Url, nil
}

func (c *Client) node(addr string) (v1.InternalCLusterClient, error) {
	c.mu.RLock()
	x, ok := c.clients[addr]
	if ok {
		c.mu.RUnlock()
		return x, nil
	}
	c.mu.RUnlock()

	conn, err := grpc.Dial(addr, c.dialOpts...)
	if err != nil {
		return nil, err
	}
	k := &Klient{
		ClientConn:            conn,
		InternalCLusterClient: v1.NewInternalCLusterClient(conn),
	}
	c.mu.Lock()
	c.clients[addr] = k
	return k, nil

}

func (c *Client) SendData(ctx context.Context, req *v1.Data, nodeAddr string, creds *v1.Credentials) error {
	remote, err := c.node(nodeAddr)
	if err != nil {
		return err
	}
	_, err = remote.SendData(ctx, req, c.callOpts(creds)...)
	return err
}

func (c *Client) callOpts(cred *v1.Credentials) (o []grpc.CallOption) {
	return []grpc.CallOption{
		grpc.PerRPCCredentials(&callCredential{
			insecure: c.insecure,
			creds:    cred,
		}),
	}
}

type callCredential struct {
	insecure bool
	creds    *v1.Credentials
}

var _ credentials.PerRPCCredentials = (*callCredential)(nil)

func (c *callCredential) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"username": c.creds.GetUsername(),
		"password": c.creds.GetPassword(),
	}, nil
}

func (c *callCredential) RequireTransportSecurity() bool {
	return !c.insecure
}

func (c *Client) Join(ctx context.Context, req *v1.Join_Request, nodeAddr string, creds *v1.Credentials) error {
	remote, err := c.node(nodeAddr)
	if err != nil {
		return err
	}
	_, err = remote.Join(ctx, req, c.callOpts(creds)...)
	return err
}

func (c *Client) Backup(ctx context.Context, w io.Writer, req *v1.Backup_Request, nodeAddr string, creds *v1.Credentials) error {
	remote, err := c.node(nodeAddr)
	if err != nil {
		return err
	}
	res, err := remote.Backup(ctx, req, c.callOpts(creds)...)
	if err != nil {
		return err
	}
	for {
		data, err := res.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		_, err = w.Write(data.Data)
		if err != nil {
			return err
		}
	}
}

func (c *Client) Load(ctx context.Context, req *v1.Load_Request, nodeAddr string, creds *v1.Credentials) error {
	remote, err := c.node(nodeAddr)
	if err != nil {
		return err
	}
	_, err = remote.Load(ctx, req, c.callOpts(creds)...)
	return err
}

func (c *Client) Notify(ctx context.Context, req *v1.Notify_Request, nodeAddr string, creds *v1.Credentials) error {
	remote, err := c.node(nodeAddr)
	if err != nil {
		return err
	}
	_, err = remote.Notify(ctx, req, c.callOpts(creds)...)
	return err
}

func (c *Client) RemoveNode(ctx context.Context, req *v1.RemoveNode_Request, nodeAddr string, creds *v1.Credentials) error {
	remote, err := c.node(nodeAddr)
	if err != nil {
		return err
	}
	_, err = remote.RemoveNode(ctx, req, c.callOpts(creds)...)
	return err
}

func (c *Client) Realtime(ctx context.Context, req *v1.Realtime_Request, nodeAddr string, creds *v1.Credentials) (*v1.Realtime_Response, error) {
	remote, err := c.node(nodeAddr)
	if err != nil {
		return nil, err
	}
	return remote.Realtime(ctx, req, c.callOpts(creds)...)
}

func (c *Client) Aggregate(ctx context.Context, req *v1.Aggregate_Request, nodeAddr string, creds *v1.Credentials) (*v1.Aggregate_Response, error) {
	remote, err := c.node(nodeAddr)
	if err != nil {
		return nil, err
	}
	return remote.Aggregate(ctx, req, c.callOpts(creds)...)
}

func (c *Client) Timeseries(ctx context.Context, req *v1.Timeseries_Request, nodeAddr string, creds *v1.Credentials) (*v1.Timeseries_Response, error) {
	remote, err := c.node(nodeAddr)
	if err != nil {
		return nil, err
	}
	return remote.Timeseries(ctx, req, c.callOpts(creds)...)
}

func (c *Client) Breakdown(ctx context.Context, req *v1.BreakDown_Request, nodeAddr string, creds *v1.Credentials) (*v1.BreakDown_Response, error) {
	remote, err := c.node(nodeAddr)
	if err != nil {
		return nil, err
	}
	return remote.BreakDown(ctx, req, c.callOpts(creds)...)
}

// Stats returns stats on the Client instance
func (c *Client) Status() *v1.Status_Cluster {
	c.mu.RLock()
	defer c.mu.RUnlock()

	o := &v1.Status_Cluster{
		LocalNodeAddress: c.localNodeAddr,
	}
	return o
}
