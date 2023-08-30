package vince

import (
	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/vince/internal/cmd/bench"
	"github.com/vinceanalytics/vince/internal/cmd/cluster"
	"github.com/vinceanalytics/vince/internal/cmd/login"
	"github.com/vinceanalytics/vince/internal/cmd/query"
	"github.com/vinceanalytics/vince/internal/cmd/serve"
	"github.com/vinceanalytics/vince/internal/cmd/sites"
	"github.com/vinceanalytics/vince/internal/cmd/vinit"
	"github.com/vinceanalytics/vince/internal/v8s"
	"github.com/vinceanalytics/vince/internal/version"
)

func App() *cli.App {
	return &cli.App{
		Name:    "vince",
		Usage:   "The Cloud Native Web Analytics Platform.",
		Version: version.Build().String(),
		Authors: []any{
			"Geofrey Ernest",
		},
		Copyright: "@2033 - present",
		Commands: []*cli.Command{
			vinit.CMD(),
			serve.CMD(),
			login.CMD(),
			sites.CMD(),
			query.CMD(),
			cluster.CMD(),
			v8s.CMD(),
			bench.CMD(),
		},
		EnableShellCompletion: true,
	}
}
