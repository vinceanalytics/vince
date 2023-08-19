package run

import (
	"os"

	"github.com/urfave/cli/v3"
)

func Main(app *cli.App) {

	err := app.Run(os.Args)
	if err != nil {
		os.Exit(1)
	}
}
