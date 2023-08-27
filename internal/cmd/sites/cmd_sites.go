package sites

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/vince/internal/cmd/ansi"
	"github.com/vinceanalytics/vince/internal/cmd/auth"
	"github.com/vinceanalytics/vince/internal/klient"
	v1 "github.com/vinceanalytics/vince/proto/v1"
)

func CMD() *cli.Command {
	return &cli.Command{
		Name:  "sites",
		Usage: "Manage sites",
		Commands: []*cli.Command{
			create(),
			list(),
			del(),
		},
	}
}

func create() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Creates a new site",
		Action: func(ctx *cli.Context) error {
			name := ctx.Args().First()
			if name == "" {
				ansi.Err("missing site domain")
				ansi.Suggestion(
					"vince sites create vinceanalytics.github.io",
				)
				os.Exit(1)
			}
			token, instance := auth.Account()

			klient.CLI_POST(
				context.Background(),
				instance+"/sites",
				&v1.Site_CreateOptions{Domain: name},
				&v1.Site{},
				token,
			)
			ansi.Ok("ok")
			return nil
		},
	}
}

func del() *cli.Command {
	return &cli.Command{
		Name:  "delete",
		Usage: "Deletes a  site",
		Action: func(ctx *cli.Context) error {
			name := ctx.Args().First()
			if name == "" {
				ansi.Err("missing site domain")
				ansi.Suggestion(
					"vince sites delete vinceanalytics.github.io",
				)
				os.Exit(1)
			}
			token, instance := auth.Account()

			klient.CLI_DELETE(
				context.Background(),
				instance+"/sites",
				&v1.Site_DeleteOptions{Domain: name},
				&v1.Site{},
				token,
			)
			ansi.Ok("ok")
			return nil
		},
	}
}

func list() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "Lists  sites",
		Action: func(ctx *cli.Context) error {
			token, instance := auth.Account()
			var list v1.Site_List
			klient.CLI_GET(
				context.Background(),
				instance+"/sites",
				&v1.Site_ListOptions{},
				&list,
				token,
			)
			for _, s := range list.List {
				ansi.Ok(s.Domain)
			}
			return nil
		},
	}
}
