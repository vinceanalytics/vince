package main

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
	"time"

	"github.com/apache/arrow/go/v15/arrow/memory"
	"github.com/bufbuild/protovalidate-go"
	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/vince/api"
	"github.com/vinceanalytics/vince/db"
	v1 "github.com/vinceanalytics/vince/gen/go/staples/v1"
	"github.com/vinceanalytics/vince/geo"
	"github.com/vinceanalytics/vince/guard"
	"github.com/vinceanalytics/vince/index/primary"
	"github.com/vinceanalytics/vince/load"
	"github.com/vinceanalytics/vince/logger"
	"github.com/vinceanalytics/vince/lsm"
	"github.com/vinceanalytics/vince/request"
	"github.com/vinceanalytics/vince/session"
	"github.com/vinceanalytics/vince/staples"
	"github.com/vinceanalytics/vince/version"
	"golang.org/x/crypto/acme/autocert"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

func main() {
	cmd := app()
	err := cmd.Run(context.Background(), os.Args)
	if err != nil {
		logger.Fail("Exited process", "err", err)
	}
}

func app() *cli.Command {
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
				Domains:         c.StringSlice("domains"),
				RetentionPeriod: durationpb.New(c.Duration("retentionPeriod")),
				AutoTls:         c.Bool("autoTls"),
			}
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
			valid, err := protovalidate.New()
			if err != nil {
				return err
			}
			err = valid.Validate(base)
			if err != nil {
				return err
			}

			log := slog.Default()
			log.Info("Setup storage")

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
			store, err := db.NewKV(base.Data)
			if err != nil {
				return err
			}

			log.Info("Setup session")
			alloc := memory.NewGoAllocator()
			log.Info("Loading primary index")
			pidx, err := primary.NewPrimary(store)
			if err != nil {
				return err
			}
			idx := staples.NewIndex()
			sess := session.New(alloc, "staples", store, idx, pidx,
				lsm.WithTTL(
					base.RetentionPeriod.AsDuration(),
				),
				lsm.WithCompactSize(
					uint64(base.GranuleSize),
				),
			)
			ctx = session.With(ctx, sess)
			log.Info("Setup geo ip")
			gip := geo.Open(base.GeoipDbPath)
			if base.GeoipDbPath == "" {
				log.Warn("Skipping geo ip, missing database path")
			}
			ctx = geo.With(ctx, gip)
			log.Info("Setup guard", "rate-limit", base.RateLimit)
			gd := guard.New(base)
			ctx = guard.With(ctx, gd)

			log.Info("Setup api")
			validate, err := protovalidate.New()
			if err != nil {
				return err
			}
			ctx = request.With(ctx, validate)
			a, err := api.New(ctx, base)
			if err != nil {
				return err
			}
			ctx = logger.With(ctx, log)
			ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
			defer cancel()

			svr := &http.Server{
				Addr:        base.Listen,
				Handler:     a,
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
			// start services
			sess.Start(ctx)
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
			sess.Close()
			return err
		},
	}
}
