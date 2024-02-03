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

	"github.com/apache/arrow/go/v15/arrow/memory"
	"github.com/bufbuild/protovalidate-go"
	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/staples/staples/api"
	"github.com/vinceanalytics/staples/staples/db"
	v1 "github.com/vinceanalytics/staples/staples/gen/go/staples/v1"
	"github.com/vinceanalytics/staples/staples/geo"
	"github.com/vinceanalytics/staples/staples/guard"
	"github.com/vinceanalytics/staples/staples/index/primary"
	"github.com/vinceanalytics/staples/staples/logger"
	"github.com/vinceanalytics/staples/staples/session"
	"github.com/vinceanalytics/staples/staples/staples"
	"github.com/vinceanalytics/staples/staples/version"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
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
		Name:      "staples",
		Usage:     "API first Cloud Native Web Analytics For Startups",
		Copyright: "@2024-present",
		Version:   version.VERSION,
		Authors: []any{
			"Geofrey Ernest",
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "data",
				Usage: "Path to store data",
				Value: "staples-data",
			},
			&cli.StringFlag{
				Name:  "listen",
				Usage: "HTTP address to listen",
				Value: ":8080",
			},
			&cli.FloatFlag{
				Name:  "rate-limit",
				Usage: "Rate limit requests",
				Value: math.MaxFloat64,
			},
			&cli.IntFlag{
				Name:  "granule-size",
				Usage: "Maximum size of block to persist",
				Value: 256 << 20, //256 MB
			},
			&cli.StringFlag{
				Name:  "geoip-db",
				Usage: "Path to geo ip database file",
			},
			&cli.StringSliceFlag{
				Name:  "domains",
				Usage: "Domain names to accept",
			},
			&cli.StringFlag{
				Name:  "config",
				Usage: "Path to configuration file",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			base := &v1.Config{
				Data:        c.String("data"),
				Listen:      c.String("listen"),
				RateLimit:   c.Float("rate-limit"),
				GranuleSize: c.Int("granule-size"),
				GeoipDbPath: c.String("geoip-db"),
				Domains:     c.StringSlice("domains"),
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
			sess := session.New(alloc, "staples", store, idx, pidx)
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
			go func() {
				defer cancel()
				log.Info("starting server", "addr", base.Listen)
				err = svr.ListenAndServe()
			}()
			<-ctx.Done()
			svr.Shutdown(context.Background())
			return err
		},
	}
}
