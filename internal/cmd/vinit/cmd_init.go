package vinit

import "github.com/urfave/cli/v3"

func CMD() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Initializes a vince project",
		Action: func(ctx *cli.Context) error {
			return nil
		},
	}
}
