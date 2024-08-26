package main

import (
	"context"
	"log"
	"os"

	"github.com/vinceanalytics/vince/internal/cmd"
)

func main() {
	err := cmd.Cli().Run(context.Background(), os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
