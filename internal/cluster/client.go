package cluster

import (
	"context"
	"errors"
	"io"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/cluster/auth"
	"github.com/vinceanalytics/vince/internal/cluster/connections"
	"github.com/vinceanalytics/vince/internal/cluster/http"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

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
	insecure bool
	conns    *connections.Manager
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
func NewClient(mgr *connections.Manager) *Client {
	return &Client{
		conns: mgr,
	}
}

// GetNodeAPIAddr retrieves the API Address for the node at nodeAddr
func (c *Client) GetNodeAPIAddr(ctx context.Context, nodeAddr string) (string, error) {
	remote, err := c.node(nodeAddr)
	if err != nil {
		return "", err
	}
	meta, err := remote.NodeAPI(ctx, &v1.NodeAPIRequest{})
	if err != nil {
		return "", err
	}
	return meta.Url, nil
}

func (c *Client) node(addr string) (v1.InternalCLusterClient, error) {
	return c.conns.ByAddress(addr)
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
		"vince_username": c.creds.GetUsername(),
		"vince_password": c.creds.GetPassword(),
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
	o := &v1.Status_Cluster{
		LocalNodeAddress: c.conns.LocalAddress(),
	}
	return o
}
