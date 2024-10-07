package cmd

import (
	"context"

	"github.com/urfave/cli/v3"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/version"
)

func Cli() *cli.Command {
	config := v1.Config{
		Admin: &v1.User{},
	}
	return &cli.Command{
		Name:        "vince",
		Usage:       "The cloud native web analytics server",
		Description: `Self hosted web analytics server that respects user privacy`,
		Version:     version.VERSION,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "listen",
				Usage:       "host:port to dind the servser",
				Value:       ":8080",
				Sources:     cli.EnvVars("VINCE_LISTEN"),
				Destination: &config.Listen,
			},
			&cli.StringFlag{
				Name:        "data",
				Usage:       "directory to store data",
				Sources:     cli.EnvVars("VINCE_DATA"),
				Destination: &config.DataPath,
			},
			&cli.BoolFlag{
				Name:        "autoTLS",
				Usage:       "enables automatic tls",
				Sources:     cli.EnvVars("VINCE_AUTO_TLS"),
				Destination: &config.Acme,
			},
			&cli.StringFlag{
				Name:        "acmeEmail",
				Usage:       "email address for atomatic tls",
				Sources:     cli.EnvVars("VINCE_ACME_EMAIL"),
				Destination: &config.AcmeEmail,
			},
			&cli.StringFlag{
				Name:        "acmeDomain",
				Usage:       "domain for atomatic tls",
				Sources:     cli.EnvVars("VINCE_ACME_DOMAIN"),
				Destination: &config.AcmeDomain,
			},
			&cli.StringFlag{
				Name:        "adminName",
				Usage:       "admin user name",
				Sources:     cli.EnvVars("VINCE_ADMIN_NAME"),
				Destination: &config.Admin.Name,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "adminEmail",
				Usage:       "admin email address",
				Sources:     cli.EnvVars("VINCE_ADMIN_EMAIL"),
				Destination: &config.Admin.Email,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "adminPassword",
				Usage:       "admin password",
				Sources:     cli.EnvVars("VINCE_ADMIN_PASSWORD"),
				Destination: &config.Admin.Password,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "url",
				Value:       "http://localhost:8080",
				Usage:       "url resolving to this vince instance",
				Sources:     cli.EnvVars("VINCE_URL"),
				Destination: &config.Url,
			},
			&cli.StringSliceFlag{
				Name:        "domains",
				Usage:       "list of domains to create on startup",
				Sources:     cli.EnvVars("VINCE_DOMAINS"),
				Destination: &config.Domains,
			},
			&cli.BoolFlag{
				Name:        "profile",
				Usage:       "registrer http profiles on /debug/ path",
				Sources:     cli.EnvVars("VINCE_PROFILE"),
				Destination: &config.Profile,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			run(&config)
			return nil
		},
	}
}
