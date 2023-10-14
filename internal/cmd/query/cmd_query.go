package query

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
	v1 "github.com/vinceanalytics/proto/gen/go/vince/query/v1"
	"github.com/vinceanalytics/vince/internal/cmd/ansi"
	"github.com/vinceanalytics/vince/internal/cmd/auth"
	"github.com/vinceanalytics/vince/internal/cmd/output"
	"github.com/vinceanalytics/vince/internal/cmd/queryfmt"
	"github.com/vinceanalytics/vince/internal/do"
)

func CMD() *cli.Command {
	return &cli.Command{
		Name:  "query",
		Usage: "connect to vince and execute sql query",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "show-query",
				Aliases: []string{"s"},
			},
		},
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
			if ctx.Bool("show-query") {
				queryfmt.Format(os.Stdout, string(file))
				fmt.Fprint(os.Stdout, "-----------")
			}
			fmt.Fprintln(os.Stdout)
			return output.Tab(os.Stdout, result)
		},
	}
}
