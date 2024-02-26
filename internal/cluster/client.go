package cluster

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/cluster/auth"
	"github.com/vinceanalytics/vince/internal/cluster/rtls"
	"github.com/vinceanalytics/vince/internal/cluster/tcp"
	"github.com/vinceanalytics/vince/internal/cluster/tcp/pool"
	"google.golang.org/protobuf/proto"
	pb "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

const (
	initialPoolSize   = 4
	maxPoolCapacity   = 64
	defaultMaxRetries = 8

	protoBufferLengthSize = 8
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
	dialer  Dialer
	timeout time.Duration

	lMu           sync.RWMutex
	localNodeAddr string
	localServ     *Service

	mu            sync.RWMutex
	poolInitialSz int
	pools         map[string]pool.Pool
}

// NewClient returns a client instance for talking to a remote node.
// Clients will retry certain commands if they fail, to allow for
// remote node restarts. Cluster management operations such as joining
// and removing nodes are not retried, to make it clear to the operator
// that the operation failed. In addition, higher-level code will
// usually retry these operations.
func NewClient(dl Dialer, t time.Duration) *Client {
	return &Client{
		dialer:        dl,
		timeout:       t,
		poolInitialSz: initialPoolSize,
		pools:         make(map[string]pool.Pool),
	}
}

// SetLocal informs the client instance of the node address for the node
// using this client. Along with the Service instance it allows this
// client to serve requests for this node locally without the network hop.
func (c *Client) SetLocal(nodeAddr string, serv *Service) error {
	c.lMu.Lock()
	defer c.lMu.Unlock()
	c.localNodeAddr = nodeAddr
	c.localServ = serv
	return nil
}

// GetNodeAPIAddr retrieves the API Address for the node at nodeAddr
func (c *Client) GetNodeAPIAddr(nodeAddr string, timeout time.Duration) (string, error) {
	c.lMu.RLock()
	defer c.lMu.RUnlock()
	if c.localNodeAddr == nodeAddr && c.localServ != nil {
		// Serve it locally!
		return c.localServ.GetNodeAPIURL(), nil
	}
	r := v1.Command_Request{
		Request: &v1.Command_Request_NodeApi{
			NodeApi: &v1.NodeAPI_Request{},
		},
	}
	var w v1.Command_Response
	_, err := c.retry(&w, &r, nodeAddr, timeout, defaultMaxRetries)
	if err != nil {
		return "", err
	}

	result := w.GetNodeApi()
	if result.Error != "" {
		return "", errors.New(result.Error)
	}

	return result.Url, nil
}

func (c *Client) Data(ctx context.Context, req *v1.DataService_Request, nodeAddr string, creds *v1.Credentials, timeout time.Duration, retries int) error {
	r := v1.Command_Request{
		Credentials: creds,
		Request: &v1.Command_Request_Data{
			Data: req,
		},
	}
	var w v1.Command_Response
	_, err := c.retry(&w, &r, nodeAddr, timeout, retries)
	if err != nil {
		return err
	}
	result := w.GetData()
	if result.Error != "" {
		return errors.New(result.Error)
	}
	return nil
}
func (c *Client) Join(ctx context.Context, req *v1.Join_Request, nodeAddr string, creds *v1.Credentials, timeout time.Duration, retries int) error {
	r := v1.Command_Request{
		Credentials: creds,
		Request: &v1.Command_Request_Join{
			Join: req,
		},
	}
	var w v1.Command_Response
	_, err := c.retry(&w, &r, nodeAddr, timeout, retries)
	if err != nil {
		return err
	}
	result := w.GetJoin()
	if result.Error != "" {
		return errors.New(result.Error)
	}
	return nil
}

func (c *Client) Backup(ctx context.Context, req *v1.Backup_Request, nodeAddr string, creds *v1.Credentials, timeout time.Duration, retries int) error {
	r := v1.Command_Request{
		Credentials: creds,
		Request: &v1.Command_Request_Backup{
			Backup: req,
		},
	}
	var w v1.Command_Response
	_, err := c.retry(&w, &r, nodeAddr, timeout, retries)
	if err != nil {
		return err
	}
	result := w.GetBackup()
	if result.Error != "" {
		return errors.New(result.Error)
	}
	return nil
}

func (c *Client) Load(ctx context.Context, req *v1.Load_Request, nodeAddr string, creds *v1.Credentials, timeout time.Duration, retries int) error {
	r := v1.Command_Request{
		Credentials: creds,
		Request: &v1.Command_Request_Load{
			Load: req,
		},
	}
	var w v1.Command_Response
	_, err := c.retry(&w, &r, nodeAddr, timeout, retries)
	if err != nil {
		return err
	}
	result := w.GetLoad()
	if result.Error != "" {
		return errors.New(result.Error)
	}
	return nil
}

func (c *Client) RemoveNode(ctx context.Context, req *v1.RemoveNode_Request, nodeAddr string, creds *v1.Credentials, timeout time.Duration, retries int) error {
	r := v1.Command_Request{
		Credentials: creds,
		Request: &v1.Command_Request_RemoveNode{
			RemoveNode: req,
		},
	}
	var w v1.Command_Response
	_, err := c.retry(&w, &r, nodeAddr, timeout, retries)
	if err != nil {
		return err
	}
	result := w.GetRemoveNode()
	if result.Error != "" {
		return errors.New(result.Error)
	}
	return nil
}

func (c *Client) Realtime(ctx context.Context, req *v1.Realtime_Request, nodeAddr string, creds *v1.Credentials, timeout time.Duration, retries int) (*v1.Realtime_Response, error) {
	r := v1.Command_Request{
		Credentials: creds,
		Request: &v1.Command_Request_Query{
			Query: &v1.Query_Request{
				Params: &v1.Query_Request_Realtime{
					Realtime: req,
				},
			},
		},
	}
	var w v1.Command_Response
	_, err := c.retry(&w, &r, nodeAddr, timeout, retries)
	if err != nil {
		return nil, err
	}
	result := w.GetQuery()
	if result.Error != "" {
		return nil, errors.New(result.Error)
	}
	return result.GetRealtime(), nil
}

func (c *Client) Aggregate(ctx context.Context, req *v1.Aggregate_Request, nodeAddr string, creds *v1.Credentials, timeout time.Duration, retries int) (*v1.Aggregate_Response, error) {
	r := v1.Command_Request{
		Credentials: creds,
		Request: &v1.Command_Request_Query{
			Query: &v1.Query_Request{
				Params: &v1.Query_Request_Aggregate{
					Aggregate: req,
				},
			},
		},
	}
	var w v1.Command_Response
	_, err := c.retry(&w, &r, nodeAddr, timeout, retries)
	if err != nil {
		return nil, err
	}
	result := w.GetQuery()
	if result.Error != "" {
		return nil, errors.New(result.Error)
	}
	return result.GetAggregate(), nil
}

func (c *Client) Timeseries(ctx context.Context, req *v1.Timeseries_Request, nodeAddr string, creds *v1.Credentials, timeout time.Duration, retries int) (*v1.Timeseries_Response, error) {
	r := v1.Command_Request{
		Credentials: creds,
		Request: &v1.Command_Request_Query{
			Query: &v1.Query_Request{
				Params: &v1.Query_Request_Timeseries{
					Timeseries: req,
				},
			},
		},
	}
	var w v1.Command_Response
	_, err := c.retry(&w, &r, nodeAddr, timeout, retries)
	if err != nil {
		return nil, err
	}
	result := w.GetQuery()
	if result.Error != "" {
		return nil, errors.New(result.Error)
	}
	return result.GetTimeseries(), nil
}
func (c *Client) Breakdown(ctx context.Context, req *v1.BreakDown_Request, nodeAddr string, creds *v1.Credentials, timeout time.Duration, retries int) (*v1.BreakDown_Response, error) {
	r := v1.Command_Request{
		Credentials: creds,
		Request: &v1.Command_Request_Query{
			Query: &v1.Query_Request{
				Params: &v1.Query_Request_Breakdown{
					Breakdown: req,
				},
			},
		},
	}
	var w v1.Command_Response
	_, err := c.retry(&w, &r, nodeAddr, timeout, retries)
	if err != nil {
		return nil, err
	}
	result := w.GetQuery()
	if result.Error != "" {
		return nil, errors.New(result.Error)
	}
	return result.GetBreakdown(), nil
}

// Stats returns stats on the Client instance
func (c *Client) Stats() *v1.Status_Cluster {
	c.mu.RLock()
	defer c.mu.RUnlock()

	o := &v1.Status_Cluster{
		Timeout:          durationpb.New(c.timeout),
		LocalNodeAddress: c.localNodeAddr,
	}
	return o
}

func (c *Client) dial(nodeAddr string, timeout time.Duration) (net.Conn, error) {
	var pl pool.Pool
	var ok bool

	c.mu.RLock()
	pl, ok = c.pools[nodeAddr]
	c.mu.RUnlock()

	// Do we need a new pool for the given address?
	if !ok {
		if err := func() error {
			c.mu.Lock()
			defer c.mu.Unlock()
			pl, ok = c.pools[nodeAddr]
			if ok {
				return nil // Pool was inserted just after we checked.
			}

			// New pool is needed for given address.
			factory := func() (net.Conn, error) { return c.dialer.Dial(nodeAddr, c.timeout) }
			p, err := pool.NewChannelPool(c.poolInitialSz, maxPoolCapacity, factory)
			if err != nil {
				return err
			}
			c.pools[nodeAddr] = p
			pl = p
			return nil
		}(); err != nil {
			return nil, err
		}
	}

	// Got pool, now get a connection.
	conn, err := pl.Get()
	if err != nil {
		return nil, fmt.Errorf("pool get: %w", err)
	}
	return conn, nil
}

// retry retries a command on a remote node. It does this so we churn through connections
// in the pool if we hit an error, as the remote node may have restarted and the pool's
// connections are now stale.
func (c *Client) retry(w *v1.Command_Response, r *v1.Command_Request, nodeAddr string, timeout time.Duration, maxRetries int) (int, error) {
	var errOuter error
	var nRetries int
	for {
		errOuter = func() error {
			conn, errInner := c.dial(nodeAddr, c.timeout)
			if errInner != nil {
				return errInner
			}
			defer conn.Close()

			if errInner = writeCommand(conn, r, timeout); errInner != nil {
				handleConnError(conn)
				return errInner
			}

			errInner = readResponse(w, conn, timeout)
			if errInner != nil {
				handleConnError(conn)
				return errInner
			}
			return nil
		}()
		if errOuter == nil {
			break
		}
		nRetries++
		if nRetries > maxRetries {
			return nRetries, errOuter
		}
	}
	return nRetries, nil
}

func writeCommand(conn net.Conn, c *v1.Command_Request, timeout time.Duration) error {
	p, err := pb.Marshal(c)
	if err != nil {
		return fmt.Errorf("command marshal: %w", err)
	}

	// Write length of Protobuf
	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}
	b := make([]byte, protoBufferLengthSize)
	binary.LittleEndian.PutUint64(b[0:], uint64(len(p)))
	_, err = conn.Write(b)
	if err != nil {
		return fmt.Errorf("write length: %w", err)
	}
	// Write actual protobuf.
	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}
	_, err = conn.Write(p)
	if err != nil {
		return fmt.Errorf("write protobuf bytes: %w", err)
	}
	return nil
}

func readResponse(w *v1.Command_Response, conn net.Conn, timeout time.Duration) (retErr error) {
	defer func() {
		// Connecting to an open port, but not a rqlite Raft API, may cause a panic
		// when the system tries to read the response. This is a workaround.
		if r := recover(); r != nil {
			retErr = fmt.Errorf("panic reading response from node: %v", r)
		}
	}()

	// Read length of incoming response.
	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}
	b := make([]byte, protoBufferLengthSize)
	_, err := io.ReadFull(conn, b)
	if err != nil {
		return fmt.Errorf("read protobuf length: %w", err)
	}
	sz := binary.LittleEndian.Uint64(b[0:])

	// Read in the actual response.
	p := make([]byte, sz)
	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}
	_, err = io.ReadFull(conn, p)
	if err != nil {
		return fmt.Errorf("read protobuf bytes: %w", err)
	}
	var resp v1.Command_Response
	err = proto.Unmarshal(p, &resp)
	if err != nil {
		return fmt.Errorf("decode protobuf bytes: %w", err)
	}
	return nil
}

func handleConnError(conn net.Conn) {
	if pc, ok := conn.(*pool.Conn); ok {
		pc.MarkUnusable()
	}
}
