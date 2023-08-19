package sites

import (
	"context"
	"errors"
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
			token, instance := account()

			err := klient.POST(
				context.Background(),
				instance+"/sites",
				&v1.Site_CreateOptions{Domain: name},
				&v1.Site{},
				token,
			)
			if err != nil {
				return ansi.ERROR(errors.New(err.Error))
			}
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
			token, instance := account()

			err := klient.DELETE(
				context.Background(),
				instance+"/sites",
				&v1.Site_DeleteOptions{Domain: name},
				&v1.Site{},
				token,
			)
			if err != nil {
				return ansi.ERROR(errors.New(err.Error))
			}
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
			token, instance := account()
			var list v1.Site_List
			err := klient.GET(
				context.Background(),
				instance+"/sites",
				&v1.Site_ListOptions{},
				&list,
				token,
			)
			if err != nil {
				return ansi.ERROR(errors.New(err.Error))
			}
			for _, s := range list.List {
				ansi.Ok(s.Domain)
			}
			return nil
		},
	}
}

func account() (token string, api string) {
	o, _ := auth.LoadClient()
	if o.Active == nil {
		ansi.Err("no active account found")
		ansi.Suggestion(
			"log in to a vince instance with [vince login] command",
			"select existing vince instance/account using [vince use] command",
		)
		os.Exit(1)
	}
	token = o.Instance[o.Active.Instance].Accounts[o.Active.Account].Token
	api = o.Active.Instance
	return
}
