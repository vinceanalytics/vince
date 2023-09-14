package query

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/query/v1"
	"github.com/vinceanalytics/vince/internal/cmd/ansi"
	"github.com/vinceanalytics/vince/internal/cmd/auth"
	"github.com/vinceanalytics/vince/internal/cmd/output"
	"github.com/vinceanalytics/vince/internal/do"
)

func CMD() *cli.Command {
	return &cli.Command{
		Name:  "query",
		Usage: "connect to vince and execute sql query",
		Action: func(ctx *cli.Context) error {
			w := ansi.New()
			a := ctx.Args().First()
			if a == "" {
				return nil
			}
			file, err := os.ReadFile(a)
			if err != nil {
				return w.Complete(err)
			}
			token, instance := auth.Account()

			result, err := do.Query(context.TODO(),
				instance, token, &v1.QueryRequest{Query: string(file)},
			)
			if err != nil {
				return w.Complete(err)
			}
			return output.Tab(os.Stdout, result)
		},
	}
}
