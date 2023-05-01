package run

import (
	"fmt"
	"os"

	"github.com/gernest/vince/cmd/man"
	"github.com/gernest/vince/config"
	"github.com/gernest/vince/pkg/version"
	"github.com/gernest/vince/server"
	"github.com/urfave/cli/v3"
)

func Main() {
	app := &cli.App{
		Name:  "vince",
		Usage: "The Cloud Native Web Analytics Platform.",
		Flags: config.Flags(),
		Commands: []*cli.Command{
			config.ConfigCMD(),
			version.Version(),
			man.Page(),
		},
		EnableShellCompletion: true,
		Action:                server.Serve,
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
