package config

import (
	"os"
	"time"

	"log/slog"

	"github.com/urfave/cli/v3"
	v1 "github.com/vinceanalytics/vince/proto/v1"
	"google.golang.org/protobuf/types/known/durationpb"
)

const (
	FILE        = "vince.json"
	META_PATH   = "meta"
	BLOCKS_PATH = "blocks"
)

var (
	DefaultSyncInterval = durationpb.New(time.Minute)
)

type Options = v1.Config

func Defaults() *v1.Config {
	return &v1.Config{
		ListenAddress:      ":8080",
		LogLevel:           "debug",
		MetaPath:           META_PATH,
		BlocksPath:         BLOCKS_PATH,
		SyncInterval:       DefaultSyncInterval,
		MysqlListenAddress: ":3306",
	}
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
			Destination: &o.TlsCertFile,
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
	}
}

func IsTLS(o *Options) bool {
	return o.TlsCertFile != "" &&
		o.TlsKeyFile != ""
}
