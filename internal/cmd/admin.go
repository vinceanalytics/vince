package cmd

import (
	"context"
	"log/slog"

	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/vince/internal/ops"
	"github.com/vinceanalytics/vince/internal/shards"
)

var admin = &cli.Command{
	Name:  "admin",
	Usage: "Creates admin account",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "data",
			Usage:   "directory to store data",
			Sources: cli.EnvVars("VINCE_DATA"),
			Value:   "vince-data",
		},
		&cli.StringFlag{
			Name:     "name",
			Usage:    "administrator name",
			Sources:  cli.EnvVars("VINCE_ADMIN_NAME"),
			Required: true,
		},
		&cli.StringFlag{
			Name:     "password",
			Usage:    "administrator password",
			Sources:  cli.EnvVars("VINCE_ADMIN_PASSWORD"),
			Required: true,
		},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		dataPath := c.String("data")
		db, err := shards.New(dataPath)
		if err != nil {
			return err
		}
		defer db.Close()

		err = ops.CreateAdmin(db.Get(), c.String("name"), c.String("password"))
		if err != nil {
			return err
		}
		a, err := ops.LoadAdmin(db.Get())
		if err != nil {
			return err
		}
		slog.Info("successfully created admin account", "name", a.Name)
		return nil
	},
}
