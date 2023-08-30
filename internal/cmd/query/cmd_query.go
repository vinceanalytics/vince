package query

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/vince/internal/cmd/ansi"
	"github.com/vinceanalytics/vince/internal/cmd/auth"
	"github.com/vinceanalytics/vince/internal/cmd/output"
	"github.com/vinceanalytics/vince/internal/klient"
	v1 "github.com/vinceanalytics/vince/proto/v1"
)

func CMD() *cli.Command {
	return &cli.Command{
		Name:  "query",
		Usage: "connect to vince and execute sql query",
		Action: func(ctx *cli.Context) error {
			a := ctx.Args().First()
			if a == "" {
				return nil
			}
			token, instance := auth.Account()
			var result v1.Query_Response
			err := klient.POST(
				context.Background(),
				instance+"/query",
				&v1.Query_Request{
					Query: a,
				},
				&result,
				token,
			)
			if err != nil {
				ansi.New().Err(err.Error).Exit()
			}
			return output.Tab(os.Stdout, &result)
		},
	}
}
