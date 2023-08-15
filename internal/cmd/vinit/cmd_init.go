package vinit

import (
	"os"

	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/internal/pj"
)

func CMD() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Initializes a vince project",
		Action: func(ctx *cli.Context) error {
			must.One(os.Mkdir(config.META_PATH, 0755))(
				"failed to create metadata directory",
			)
			must.One(os.Mkdir(config.BLOCKS_PATH, 0755))(
				"failed to create blocks directory",
			)
			b := must.Must(pj.MarshalIndent(config.Defaults()))()
			must.One(os.WriteFile(config.FILE, b, 0600))(
				"failed to create vince configuration",
			)
			return nil
		},
	}
}
