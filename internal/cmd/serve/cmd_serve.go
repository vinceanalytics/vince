package serve

import (
	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/server"
)

func CMD() *cli.Command {
	o := config.Options{}
	return &cli.Command{
		Name:  "serve",
		Usage: "Serves web ui console and expose /api/events that collects web analytics",
		Flags: config.Flags(&o),
		Action: func(ctx *cli.Context) error {
			return server.Serve(&o, ctx)
		},
	}
}
