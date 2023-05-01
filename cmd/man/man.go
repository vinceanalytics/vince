package man

import (
	"os"

	"github.com/urfave/cli/v3"
)

func Page() *cli.Command {
	return &cli.Command{
		Name:  "man",
		Usage: "Generate man pages",
		Action: func(ctx *cli.Context) error {
			m, err := ctx.App.ToMan()
			if err != nil {
				return err
			}
			path := ctx.Args().First()
			if path == "" {
				os.Stdout.WriteString(m)
				return nil
			}
			return os.WriteFile(path, []byte(m), 0600)
		},
	}
}
