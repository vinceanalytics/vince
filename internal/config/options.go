package config

import (
	"os"
	"time"

	"log/slog"

	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/vince/internal/ng"
	v1 "github.com/vinceanalytics/vince/proto/v1"
	"google.golang.org/protobuf/types/known/durationpb"
)

const (
	FILE        = "vince.json"
	META_PATH   = "meta"
	BLOCKS_PATH = "blocks"
)

var (
	DefaultSyncInterval     = durationpb.New(time.Minute)
	DefaultEventsBufferSize = 10 << 10
)

type Options = v1.Config

func Defaults() *v1.Config {
	o := &v1.Config{
		ListenAddress:      ":8080",
		LogLevel:           "debug",
		MetaPath:           META_PATH,
		BlocksPath:         BLOCKS_PATH,
		SyncInterval:       DefaultSyncInterval,
		MysqlListenAddress: ":3306",
		EventsBufferSize:   int64(DefaultEventsBufferSize),
		ServerId:           ng.Name(),
		RaftListenAddress:  ":6800",
	}
	return o
}

func Logger(level string) *slog.Logger {
	var lvl slog.Level
	lvl.UnmarshalText([]byte(level))
	return slog.New(slog.NewTextHandler(
		os.Stdout, &slog.HandlerOptions{
			Level: lvl,
		},
	))
}

func Flags(o *Options) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Category:    "core",
			Name:        "listen",
			Usage:       "http address to listen to",
			Value:       ":8080",
			Destination: &o.ListenAddress,
			EnvVars:     []string{"VINCE_LISTEN"},
		},
		&cli.StringFlag{
			Category:    "core",
			Name:        "listen-mysql",
			Usage:       "serve mysql clients on this address",
			Value:       ":3306",
			Destination: &o.ListenAddress,
			EnvVars:     []string{"VINCE_MYSQL_LISTEN"},
		},
		&cli.StringFlag{
			Category:    "core",
			Name:        "tls-cert-file",
			Usage:       "path to tls certificate",
			Destination: &o.TlsCertFile,
			EnvVars:     []string{"VINCE_TLS_CERT_FILE"},
		},
		&cli.StringFlag{
			Category:    "core",
			Name:        "tls-key-file",
			Usage:       "path to tls key",
			Destination: &o.TlsKeyFile,
			EnvVars:     []string{"VINCE_TLS_KEY_FILE"},
		},
		&cli.StringFlag{
			Category:    "core",
			Name:        "log-level",
			Usage:       "log level, values are (trace,debug,info,warn,error,fatal,panic)",
			Value:       "debug",
			Destination: &o.LogLevel,
			EnvVars:     []string{"VINCE_LOG_LEVEL"},
		},

		&cli.StringFlag{
			Category:    "core",
			Name:        "meta-path",
			Usage:       "path to meta data directory",
			Value:       META_PATH,
			Destination: &o.MetaPath,
			EnvVars:     []string{"VINCE_METa_PATH"},
		},
		&cli.StringFlag{
			Category:    "core",
			Name:        "blocks-path",
			Usage:       "Path to store block files",
			Value:       BLOCKS_PATH,
			Destination: &o.BlocksPath,
			EnvVars:     []string{"VINCE_BLOCK_PATH"},
		},

		&cli.DurationFlag{
			Category: "intervals",
			Name:     "sync-interval",
			Usage:    "window for buffering timeseries in memory before saving them",
			Value:    time.Minute,
			EnvVars:  []string{"VINCE_SYNC_INTERVAL"},
		},

		&cli.BoolFlag{
			Category:    "core",
			Name:        "enable-profile",
			Usage:       "Expose /debug/pprof endpoint",
			Destination: &o.EnableProfile,
			EnvVars:     []string{"VINCE_ENABLE_PROFILE"},
		},
		&cli.Int64Flag{
			Name:        "events-buffer-size",
			Usage:       "Number of events to keep in memory before saving",
			Value:       int64(DefaultEventsBufferSize),
			Destination: &o.EventsBufferSize,
			EnvVars:     []string{"VINCE_EVENTS_BUFFER_SIZE"},
		},
		&cli.StringFlag{
			Name:        "server-id",
			Usage:       "unique id of this server in a cluster",
			Destination: &o.ServerId,
			EnvVars:     []string{"VINCE_SERVER_ID"},
		},
		&cli.StringFlag{
			Name:        "raft-listen-address",
			Usage:       "network address for cluster communication",
			Value:       ":6800",
			Destination: &o.RaftListenAddress,
			EnvVars:     []string{"VINCE_RAFT_LISTEN_ADDRESS"},
		},
	}
}

func IsTLS(o *Options) bool {
	return o.TlsCertFile != "" &&
		o.TlsKeyFile != ""
}
