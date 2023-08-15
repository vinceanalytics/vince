package vinit

import (
	"os"
	"path/filepath"

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
			root := ctx.Args().First()
			if root != "" {
				// Try to make sure the directory exists. No need to check for error here
				// because we bail out first thing when we cant write to this path.
				os.MkdirAll(root, 0755)
			}
			meta := config.META_PATH
			if root != "" {
				meta = filepath.Join(root, meta)
			}
			must.One(os.Mkdir(meta, 0755))(
				"failed to create metadata directory",
			)
			blocks := config.BLOCKS_PATH
			if root != "" {
				blocks = filepath.Join(root, blocks)
			}
			must.One(os.Mkdir(blocks, 0755))(
				"failed to create blocks directory",
			)
			b := must.Must(pj.MarshalIndent(config.Defaults()))()
			file := config.FILE
			if root != "" {
				file = filepath.Join(root, file)
			}
			must.One(os.WriteFile(file, b, 0600))(
				"failed to create vince configuration",
			)
			return nil
		},
	}
}
