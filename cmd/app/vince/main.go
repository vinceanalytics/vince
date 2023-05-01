package vince

import (
	"github.com/gernest/vince/cmd/version"
	"github.com/gernest/vince/config"
	"github.com/gernest/vince/server"
	"github.com/urfave/cli/v3"
)

func App() *cli.App {
	return &cli.App{
		Name:  "vince",
		Usage: "The Cloud Native Web Analytics Platform.",
		Flags: config.Flags(),
		Commands: []*cli.Command{
			config.ConfigCMD(),
			version.Version(),
		},
		EnableShellCompletion: true,
		Action:                server.Serve,
	}
}
