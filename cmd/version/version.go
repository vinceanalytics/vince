package version

import (
	"os"

	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/vince/pkg/version"
)

func Version() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "prints version information",
		Action: func(ctx *cli.Context) error {
			x := version.Build()
			os.Stdout.WriteString(x.String())
			return nil
		},
	}
}
