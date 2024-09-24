package cmd

import (
	"context"

	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/version"
)

func Cli() *cli.Command {
	return &cli.Command{
		Name:        "vince",
		Description: "Simple web analytics",
		Version:     version.VERSION,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "listen",
				Usage:       "host:port to dind the servser",
				Value:       ":8080",
				Sources:     cli.EnvVars("VINCE_LISTEN"),
				Destination: &config.C.Listen,
			},
			&cli.StringFlag{
				Name:        "data",
				Usage:       "directory to store data",
				Value:       ".data",
				Sources:     cli.EnvVars("VINCE_DATA"),
				Destination: &config.C.DataPath,
			},
			&cli.BoolFlag{
				Name:        "acme",
				Usage:       "enables automatic tls",
				Sources:     cli.EnvVars("VINCE_ACME"),
				Destination: &config.C.Acme,
			},
			&cli.StringFlag{
				Name:        "acmeEmail",
				Usage:       "email address for atomatic tls",
				Sources:     cli.EnvVars("VINCE_ACME_EMAIL"),
				Destination: &config.C.AcmeEmail,
			},
			&cli.StringFlag{
				Name:        "acmeDomain",
				Usage:       "domain for atomatic tls",
				Sources:     cli.EnvVars("VINCE_ACME_DOMAIN"),
				Destination: &config.C.AcmeDomain,
			},
			&cli.StringFlag{
				Name:        "adminName",
				Usage:       "admin user name",
				Sources:     cli.EnvVars("VINCE_ADMIN_NAME"),
				Destination: &config.C.Admin.Name,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "adminEmail",
				Usage:       "admin email address",
				Sources:     cli.EnvVars("VINCE_ADMIN_EMAIL"),
				Destination: &config.C.Admin.Email,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "adminPassword",
				Usage:       "admin password",
				Sources:     cli.EnvVars("VINCE_ADMIN_PASSWORD"),
				Destination: &config.C.Admin.Password,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "license",
				Value:       "",
				Usage:       "path to lincense key file",
				Sources:     cli.EnvVars("VINCE_LICENSE"),
				Required:    true,
				Destination: &config.C.License,
			},
			&cli.StringFlag{
				Name:        "url",
				Value:       "http://localhost:8080",
				Usage:       "url resolving to this vince instance",
				Sources:     cli.EnvVars("VINCE_URL"),
				Destination: &config.C.Url,
			},
			&cli.StringSliceFlag{
				Name:        "domains",
				Usage:       "list of domains to create on startup",
				Sources:     cli.EnvVars("VINCE_DOMAINS"),
				Destination: &config.C.Domains,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			run()
			return nil
		},
	}
}
