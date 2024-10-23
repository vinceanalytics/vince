package cmd

import (
	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/vince/internal/version"
)

func Cli() *cli.Command {
	return &cli.Command{
		Name:        "vince",
		Usage:       "The cloud native web analytics server",
		Description: `Self hosted web analytics server that respects user privacy`,
		Version:     version.VERSION,
		Commands:    []*cli.Command{serve, admin},
	}
}
