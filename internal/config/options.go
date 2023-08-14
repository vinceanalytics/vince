package config

import (
	"time"

	"github.com/rs/zerolog"
	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/vince/internal/must"
	v1 "github.com/vinceanalytics/vince/proto/v1"
	"google.golang.org/protobuf/types/known/durationpb"
)

var (
	DefaultSyncInterval = durationpb.New(time.Minute)
)

type Options = v1.Config

func GetLogLevel(o *Options) zerolog.Level {
	if o.LogLevel == "" {
		return zerolog.DebugLevel
	}
	return must.Must(zerolog.ParseLevel(o.LogLevel))()
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
			Value:       "meta",
			Destination: &o.MetaPath,
			EnvVars:     []string{"VINCE_METa_PATH"},
		},
		&cli.StringFlag{
			Category:    "core",
			Name:        "blocks-path",
			Usage:       "Path to store block files",
			Value:       "blocks",
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
