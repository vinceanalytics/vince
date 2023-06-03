package vince

import (
	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/server"
	"github.com/vinceanalytics/vince/pkg/version"
)

func App() *cli.App {
	o := config.Options{}
	return &cli.App{
		Name:  "vince",
		Usage: "The Cloud Native Web Analytics Platform.",
		Flags: o.Flags(),
		Commands: []*cli.Command{
			config.ConfigCMD(),
			version.VersionCmd(),
		},
		EnableShellCompletion: true,
		Action: func(ctx *cli.Context) error {
			return server.Serve(&o)
		},
	}
}
