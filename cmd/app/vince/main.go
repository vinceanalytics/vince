package vince

import (
	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/vince/cmd/version"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/server"
)

func App() *cli.App {
	o := config.Options{}
	return &cli.App{
		Name:  "vince",
		Usage: "The Cloud Native Web Analytics Platform.",
		Flags: o.Flags(),
		Commands: []*cli.Command{
			config.ConfigCMD(),
			version.Version(),
		},
		EnableShellCompletion: true,
		Action: func(ctx *cli.Context) error {
			return server.Serve(&o)
		},
	}
}
