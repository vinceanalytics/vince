package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/bufbuild/protovalidate-go"
	"github.com/urfave/cli/v3"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/cluster"
	"github.com/vinceanalytics/vince/internal/cluster/auth"
	"github.com/vinceanalytics/vince/internal/cluster/connections"
	httpd "github.com/vinceanalytics/vince/internal/cluster/http"
	"github.com/vinceanalytics/vince/internal/cluster/rtls"
	"github.com/vinceanalytics/vince/internal/cluster/store"
	"github.com/vinceanalytics/vince/internal/cluster/transport"
	"github.com/vinceanalytics/vince/internal/geo"
	"github.com/vinceanalytics/vince/internal/guard"
	"github.com/vinceanalytics/vince/internal/load"
	"github.com/vinceanalytics/vince/internal/logger"
	"github.com/vinceanalytics/vince/internal/tenant"
	"github.com/vinceanalytics/vince/version"
	"golang.org/x/crypto/acme/autocert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

func App() *cli.Command {
	return &cli.Command{
		Name:      "vince",
		Usage:     "API first high performance self hosted and cost effective privacy friendly web analytics  server for organizations of any size",
		Copyright: "@2024-present",
		Version:   version.VERSION,
		Authors: []any{
			"Geofrey Ernest",
		},
		Commands: []*cli.Command{load.CMD()},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "data",
				Usage:   "Path to store data",
				Value:   "vince-data",
				Sources: cli.EnvVars("VINCE_DATA"),
			},
			&cli.StringFlag{
				Name:    "listen",
				Usage:   "HTTP address to listen",
				Value:   ":8080",
				Sources: cli.EnvVars("VINCE_LISTEN"),
			},
			&cli.FloatFlag{
				Name:    "rateLimit",
				Usage:   "Rate limit requests",
				Value:   math.MaxFloat64,
				Sources: cli.EnvVars("VINCE_RATE_LIMIT"),
			},
			&cli.IntFlag{
				Name:    "granuleSize",
				Usage:   "Maximum size of block to persist",
				Value:   16 << 20, //256 MB
				Sources: cli.EnvVars("VINCE_GRANULE_SIZE"),
			},
			&cli.StringFlag{
				Name:    "geoipDbPath",
				Usage:   "Path to geo ip database file",
				Sources: cli.EnvVars("VINCE_GEOIP_DB"),
			},
			&cli.StringSliceFlag{
				Name:    "domains",
				Usage:   "Domain names to accept",
				Sources: cli.EnvVars("VINCE_DOMAINS"),
			},
			&cli.StringFlag{
				Name:    "config",
				Usage:   "Path to configuration file",
				Sources: cli.EnvVars("VINCE_CONFIG"),
			},
			&cli.DurationFlag{
				Name:    "retentionPeriod",
				Usage:   "How long data will be persisted",
				Value:   30 * 24 * time.Hour,
				Sources: cli.EnvVars("VINCE_RETENTION_PERIOD"),
			},
			&cli.StringFlag{
				Name:    "logLevel",
				Value:   "INFO",
				Sources: cli.EnvVars("VINCE_LOG_LEVEL"),
			},
			&cli.BoolFlag{
				Name:    "autoTls",
				Usage:   "Enables automatic tls with lets encrypt",
				Sources: cli.EnvVars("VINCE_AUTO_TLS"),
			},
			&cli.StringFlag{
				Name:    "acmeEmail",
				Sources: cli.EnvVars("VINCE_ACME_EMAIL"),
			},
			&cli.StringFlag{
				Name:    "acmeDomain",
				Sources: cli.EnvVars("VINCE_ACME_DOMAIN"),
			},
			&cli.StringFlag{
				Name:    "authToken",
				Usage:   "Bearer token to authenticate api calls",
				Sources: cli.EnvVars("VINCE_AUTH_TOKEN"),
			},
			&cli.StringFlag{
				Name:    "credentials",
				Usage:   "Path to credentials file",
				Sources: cli.EnvVars("VINCE_CREDENTIALS"),
			},
			&cli.StringFlag{
				Name:    "nodeId",
				Usage:   "Raft id of the node",
				Sources: cli.EnvVars("VINCE_NODE_ID"),
			},
			&cli.StringFlag{
				Name:    "nodeAdv",
				Usage:   "Advertised address for inter-node-communication",
				Value:   "localhost:4002",
				Sources: cli.EnvVars("VINCE_NODE_ADVERTISE"),
			},
			&cli.StringFlag{
				Name:    "nodeCa",
				Usage:   "Path to ca certificate for this node",
				Sources: cli.EnvVars("VINCE_NODE_CA"),
			},
			&cli.StringFlag{
				Name:    "nodeCert",
				Usage:   "Path to X509 certificate for this node",
				Sources: cli.EnvVars("VINCE_NODE_CERT"),
			},
			&cli.StringFlag{
				Name:    "nodeKey",
				Usage:   "Path to X509 key for this node",
				Sources: cli.EnvVars("VINCE_NODE_KEY"),
			},
			&cli.BoolFlag{
				Name:    "nodeVerify",
				Usage:   "Verify X509  certs",
				Sources: cli.EnvVars("VINCE_NODE_VERIFY"),
			},
			&cli.BoolFlag{
				Name:    "nodeVerifyCLient",
				Usage:   "Enables mutual TLS on node-to-node communications",
				Sources: cli.EnvVars("VINCE_NODE_VERIFY_CLIENT"),
			},
			&cli.BoolFlag{
				Name:    "nodeVerifyServerName",
				Usage:   "Verifies nodes host names",
				Sources: cli.EnvVars("VINCE_NODE_VERIFY_SERVER_NAME"),
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			var level slog.Level
			level.UnmarshalText([]byte(c.String("logLevel")))
			lvl := &slog.LevelVar{}
			lvl.Set(level)
			slog.SetDefault(
				slog.New(
					slog.NewJSONHandler(
						os.Stdout,
						&slog.HandlerOptions{
							Level: lvl,
						},
					),
				),
			)
			base := &v1.Config{
				Data:            c.String("data"),
				Listen:          c.String("listen"),
				RateLimit:       c.Float("rateLimit"),
				GranuleSize:     c.Int("granuleSize"),
				GeoipDbPath:     c.String("geoipDbPath"),
				RetentionPeriod: durationpb.New(c.Duration("retentionPeriod")),
				AutoTls:         c.Bool("autoTls"),
				AuthToken:       c.String("authToken"),
				Node: &v1.RaftNode{
					Id:               c.String("nodeId"),
					Advertise:        c.String("nodeAdv"),
					Ca:               c.String("nodeCa"),
					Cert:             c.String("nodeCert"),
					Key:              c.String("nodeKey"),
					Verify:           c.Bool("nodeVerify"),
					VerifyClient:     c.Bool("nodeVerifyCLient"),
					VerifyServerName: c.Bool("nodeVerifyServerName"),
				},
			}
			log := slog.Default()
			base = tenant.Config(base, c.StringSlice("domains"))
			if base.AutoTls {
				base.Acme = &v1.Acme{
					Email:  c.String("acmeEmail"),
					Domain: c.String("acmeDomain"),
				}
			}

			if co := c.String("config"); co != "" {
				d, err := os.ReadFile(co)
				if err == nil {
					var n v1.Config
					err = protojson.Unmarshal(d, &n)
					if err != nil {
						return fmt.Errorf("invalid configuration file %v", err)
					}
					proto.Merge(base, &n)
				}
			}
			if cp := c.String("credentials"); cp != "" {
				d, err := os.ReadFile(cp)
				if err != nil {
					logger.Fail("failed loading credentials file", "err", err)
				}
				var ls v1.Credential_List
				err = protojson.Unmarshal(d, &ls)
				if err != nil {
					logger.Fail("failed decoding credentials file", "err", err)
				}
				if base.Credentials == nil {
					base.Credentials = &ls
				} else {
					proto.Merge(base.Credentials, &ls)
				}
			}
			valid, err := protovalidate.New()
			if err != nil {
				return err
			}
			err = valid.Validate(base)
			if err != nil {
				return err
			}
			_, err = os.Stat(base.Data)
			if err != nil {
				if os.IsNotExist(err) {
					err = os.MkdirAll(base.Data, 0755)
					if err != nil {
						return err
					}
				} else {
					return err
				}
			}

			ctx = logger.With(ctx, log)
			ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
			defer cancel()

			// We use same listener for raft and cluster connections
			nodesConn, err := net.Listen("tcp", base.Node.Advertise)
			if err != nil {
				return fmt.Errorf("failed opening raft listener:%v", err)
			}
			connMgr := connections.New(base.Node.Advertise)
			defer connMgr.Close()
			transit := transport.New(connMgr)
			defer transit.Close()
			tenants := tenant.NewTenants(base)
			xguard := guard.New(base, tenants)

			creds := auth.NewCredentialsStore()
			if base.Credentials != nil {
				creds.Load(base.Credentials)
			}

			db, err := store.NewStore(base, transit.Transport(), connMgr, tenants)
			if err != nil {
				return err
			}
			err = db.Open(ctx)
			if err != nil {
				return err
			}
			defer db.Close()

			cluSvc := cluster.New(db)
			gSvr := grpc.NewServer(serverOptions(base.Node, creds)...)
			transit.Register(gSvr)
			v1.RegisterInternalCLusterServer(gSvr, cluSvc)
			cluCLient := cluster.NewClient(connMgr)
			xg := geo.Open(base.GeoipDbPath)
			httpSvc := httpd.New(db, cluCLient, creds, xguard, tenants, xg)

			nodes, err := db.Nodes(ctx)
			if err != nil {
				return err
			}
			err = createCluster(ctx, log, base.Node, len(nodes.GetItems()) > 0,
				cluCLient, db, httpSvc, creds)
			if err != nil {
				return fmt.Errorf("failed creating cluster %v", err)
			}
			svr := &http.Server{
				Addr:        base.Listen,
				Handler:     httpSvc,
				BaseContext: func(l net.Listener) context.Context { return ctx },
			}
			if base.AutoTls {
				m := &autocert.Manager{
					Prompt:     autocert.AcceptTOS,
					HostPolicy: autocert.HostWhitelist(base.Acme.Domain),
					Email:      base.Acme.Email,
					Cache:      autocert.DirCache(filepath.Join(base.Data, "certs")),
				}
				svr.TLSConfig = m.TLSConfig()
			}
			go func() {
				defer cancel()
				log.Info("Starting grpc server", "addr", nodesConn.Addr().String())
				err := gSvr.Serve(nodesConn)
				if err != nil {
					log.Error("exited grpc server", "err", err)
				}
			}()
			go func() {
				defer cancel()
				log.Info("starting server", "addr", base.Listen)
				if base.AutoTls {
					err = svr.ListenAndServeTLS("", "")
				} else {
					err = svr.ListenAndServe()
				}
			}()
			<-ctx.Done()
			svr.Shutdown(context.Background())
			gSvr.GracefulStop()
			nodesConn.Close()
			return err
		},
	}
}

func serverOptions(node *v1.RaftNode, creds cluster.CredentialStore) (o []grpc.ServerOption) {
	a := &cluster.Interceptor{CredentialStore: creds}
	o = []grpc.ServerOption{
		grpc.UnaryInterceptor(a.Unary),
		grpc.StreamInterceptor(a.Stream),
	}
	if node.Cert == "" || node.Key == "" {
		return nil
	}
	mTLSState := rtls.MTLSStateDisabled
	if node.VerifyClient {
		mTLSState = rtls.MTLSStateEnabled
	}
	tlsConfig, err := rtls.CreateServerConfig(node.Cert, node.Key, node.Ca, mTLSState)
	if err != nil {
		logger.Fail("Failed creating tls config for gRPC server", "err", err)
	}
	o = append(o, grpc.Creds(credentials.NewTLS(tlsConfig)))
	return
}

func createCluster(ctx context.Context, log *slog.Logger, node *v1.RaftNode, hasPeers bool, client *cluster.Client, str *store.Store, httpServ *httpd.Service, credStr *auth.CredentialsStore) error {
	if len(node.Joins) == 0 && !hasPeers {
		if node.NonVoter {
			return fmt.Errorf("cannot create a new non-voting node without joining it to an existing cluster")
		}
		// Brand new node, told to bootstrap itself. So do it.
		log.Info("bootstraping single new node")
		if err := str.Bootstrap(ctx, &v1.Server_List{
			Items: []*v1.Server{
				{Id: node.Id, Addr: node.Advertise, Suffrage: v1.Server_Voter},
			},
		}); err != nil {
			return fmt.Errorf("failed to bootstrap single new node: %s", err.Error())
		}
		return nil
	}
	return nil
}
