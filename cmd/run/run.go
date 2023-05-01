package run

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

func Main(app *cli.App) {

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
