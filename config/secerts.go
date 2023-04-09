package config

import "github.com/urfave/cli/v2"

func Secrets() *cli.Command {
	return &cli.Command{
		Name:  "secrets",
		Usage: "generates all secrets used by vince",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "path,p",
				Usage: "directory to save the secrets",
			},
		},
	}
}
