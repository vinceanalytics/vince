package vince

import (
	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/vince/internal/cmd/serve"
	"github.com/vinceanalytics/vince/internal/cmd/vinit"
	"github.com/vinceanalytics/vince/internal/v8s"
	"github.com/vinceanalytics/vince/internal/version"
)

func App() *cli.App {
	return &cli.App{
		Name:  "vince",
		Usage: "The Cloud Native Web Analytics Platform.",
		Commands: []*cli.Command{
			serve.CMD(),
			version.CMD(),
			v8s.CMD(),
			vinit.CMD(),
		},
		EnableShellCompletion: true,
	}
}
