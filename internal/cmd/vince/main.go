package vince

import (
	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/vince/internal/cmd/query"
	"github.com/vinceanalytics/vince/internal/cmd/serve"
	"github.com/vinceanalytics/vince/internal/cmd/vinit"
	"github.com/vinceanalytics/vince/internal/v8s"
	"github.com/vinceanalytics/vince/internal/version"
)

func App() *cli.App {
	return &cli.App{
		Name:    "vince",
		Usage:   "The Cloud Native Web Analytics Platform.",
		Version: version.Build().String(),
		Commands: []*cli.Command{
			serve.CMD(),
			v8s.CMD(),
			vinit.CMD(),
			query.CMD(),
		},
		EnableShellCompletion: true,
	}
}
