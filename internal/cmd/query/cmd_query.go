package query

import "github.com/urfave/cli/v3"

func CMD() *cli.Command {
	return &cli.Command{
		Name:  "query",
		Usage: "connect to vince and execute sql query",
	}
}
