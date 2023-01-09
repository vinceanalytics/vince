package main

import (
	"log"
	"os"

	"github.com/gernest/vince"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:        "vince",
		Usage:       "simple web analytics platform",
		Description: description,
		Commands: []*cli.Command{
			vince.ServeCMD(),
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

const description = `vince is a self hosted alternative to google analytics. 
Features include
	* Zero runtime dependency: Uses embedded time series database and sqlite for 
	  application data .
	* Web analytics collection.
	* Web analytics visualization.
	* Web analytics reporting. Get notified via email 
It is simple to manage and has small footprint in resource consumption. 
`
