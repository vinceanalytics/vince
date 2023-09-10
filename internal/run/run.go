package run

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"
)

func Main(app *cli.Command) {

	err := app.Run(context.Background(), os.Args)
	if err != nil {
		os.Exit(1)
	}
}
