package run

import (
	"fmt"
	"os"

	"github.com/gernest/vince/config"
	"github.com/gernest/vince/pkg/version"
	"github.com/gernest/vince/server"
	"github.com/urfave/cli/v2"
)

func Main() {
	app := &cli.App{
		Name:  "vince",
		Usage: "The open source single file, self hosted web analytics platform.",
		Flags: config.Flags(),
		Commands: []*cli.Command{
			config.ConfigCMD(),
			version.Version(),
		},
		Action: server.Serve,
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
