package config

import (
	"os"

	"log/slog"

	"github.com/urfave/cli/v3"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/config/v1"
	"github.com/vinceanalytics/vince/internal/ng"
)

const (
	FILE        = "vince.json"
	DB_PATH     = "db"
	BLOCKS_PATH = "blocks"
	RAFT_PATH   = "raft"
	SECRET_KEY  = "secret_key"
)

var (
	DefaultEventsBufferSize = 10 << 10
)

type Options = v1.Config

func Defaults() *v1.Config {
	o := &v1.Config{
		ListenAddress: ":8080",
		LogLevel:      "debug",
		DbPath:        DB_PATH,
		BlocksStore: &v1.BlockStore{
			Provider: &v1.BlockStore_Fs{
				Fs: &v1.BlockStore_Filesystem{
					Directory: BLOCKS_PATH,
				},
			},
			CacheDir: BLOCKS_PATH,
		},
		RaftPath:           RAFT_PATH,
		MysqlListenAddress: ":3306",
		EventsBufferSize:   int64(DefaultEventsBufferSize),
		ServerId:           ng.Name(),
		SecretKey: &v1.Config_SecretKey{
			Value: &v1.Config_SecretKey_File{
				File: SECRET_KEY,
			},
		},
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
			Sources:     cli.EnvVars("VINCE_LISTEN"),
		},
		&cli.StringFlag{
			Category:    "core",
			Name:        "listen-mysql",
			Usage:       "serve mysql clients on this address",
			Value:       ":3306",
			Destination: &o.ListenAddress,
			Sources:     cli.EnvVars("VINCE_MYSQL_LISTEN"),
		},
		&cli.StringFlag{
			Category:    "core",
			Name:        "tls-cert-file",
			Usage:       "path to tls certificate",
			Destination: &o.TlsCertFile,
			Sources:     cli.EnvVars("VINCE_TLS_CERT_FILE"),
		},
		&cli.StringFlag{
			Category:    "core",
			Name:        "tls-key-file",
			Usage:       "path to tls key",
			Destination: &o.TlsKeyFile,
			Sources:     cli.EnvVars("VINCE_TLS_KEY_FILE"),
		},
		&cli.StringFlag{
			Category:    "core",
			Name:        "log-level",
			Usage:       "log level, values are (trace,debug,info,warn,error,fatal,panic)",
			Value:       "debug",
			Destination: &o.LogLevel,
			Sources:     cli.EnvVars("VINCE_LOG_LEVEL"),
		},

		&cli.StringFlag{
			Category:    "core",
			Name:        "db-path",
			Usage:       "path to main database",
			Value:       DB_PATH,
			Destination: &o.DbPath,
			Sources:     cli.EnvVars("VINCE_DB_PATH"),
		},
		&cli.BoolFlag{
			Category:    "core",
			Name:        "enable-profile",
			Usage:       "Expose /debug/pprof endpoint",
			Destination: &o.EnableProfile,
			Sources:     cli.EnvVars("VINCE_ENABLE_PROFILE"),
		},
		&cli.IntFlag{
			Name:        "events-buffer-size",
			Usage:       "Number of events to keep in memory before saving",
			Value:       int64(DefaultEventsBufferSize),
			Destination: &o.EventsBufferSize,
			Sources:     cli.EnvVars("VINCE_EVENTS_BUFFER_SIZE"),
		},
		&cli.StringFlag{
			Name:        "server-id",
			Usage:       "unique id of this server in a cluster",
			Destination: &o.ServerId,
			Sources:     cli.EnvVars("VINCE_SERVER_ID"),
		},
		&cli.StringSliceFlag{
			Name:        "allowed-origins",
			Usage:       "Origins allowed for cors",
			Value:       []string{"*"},
			Destination: &o.AllowedOrigins,
			Sources:     cli.EnvVars("VINCE_ALLOWED_ORIGINS"),
		},
	}
}

func IsTLS(o *Options) bool {
	return o.TlsCertFile != "" &&
		o.TlsKeyFile != ""
}
