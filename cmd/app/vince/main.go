package vince

import (
	"github.com/gernest/vince/cmd/version"
	"github.com/gernest/vince/config"
	"github.com/gernest/vince/server"
	"github.com/urfave/cli/v3"
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
