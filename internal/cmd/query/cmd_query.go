package query

import (
	"context"
	"errors"
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
			var result v1.Query_Result
			err := klient.POST(
				context.Background(),
				instance+"/query",
				&v1.Query_RequestOptions{
					Query: a,
				},
				&result,
				token,
			)
			if err != nil {
				return ansi.ERROR(errors.New(err.Error))
			}
			return ansi.ERROR(
				output.Tab(os.Stdout, &result),
			)
		},
	}
}
