package sites

import (
	"context"
	"strings"

	"github.com/urfave/cli/v3"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/api/v1"
	"github.com/vinceanalytics/vince/internal/cmd/ansi"
	"github.com/vinceanalytics/vince/internal/cmd/auth"
	"github.com/vinceanalytics/vince/internal/klient"
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
			w := ansi.New()
			name := ctx.Args().First()
			if name == "" {
				w.Err("missing site domain")
				w.Suggest(
					"vince sites create vinceanalytics.github.io",
				).Exit()
			}
			name = strings.TrimSpace(name)
			token, instance := auth.Account()

			klient.CLI(
				context.Background(),
				instance,
				&v1.CreateSiteRequest{Domain: name},
				&v1.CreateSiteResponse{},
				token,
			)
			return w.Ok("created").Complete(nil)
		},
	}
}

func del() *cli.Command {
	return &cli.Command{
		Name:  "delete",
		Usage: "Deletes a  site",
		Action: func(ctx *cli.Context) error {
			w := ansi.New()
			name := ctx.Args().First()
			if name == "" {
				w.Err("missing site domain")
				w.Suggest(
					"vince sites delete vinceanalytics.github.io",
				).Exit()
			}
			token, instance := auth.Account()

			klient.CLI(
				context.Background(),
				instance,
				&v1.DeleteSiteRequest{Domain: name},
				&v1.DeleteSiteResponse{},
				token,
			)
			return w.Ok("deleted").Complete(nil)
		},
	}
}

func list() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "Lists  sites",
		Action: func(ctx *cli.Context) error {
			w := ansi.New()
			token, instance := auth.Account()
			var list v1.ListSitesResponse
			klient.CLI(
				context.Background(),
				instance,
				&v1.ListSitesRequest{},
				&list,
				token,
			)
			for _, s := range list.List {
				w.Ok(s.Domain)
			}
			return w.Complete(nil)
		},
	}
}
