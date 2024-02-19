package main

import (
	"context"
	"os"

	"github.com/vinceanalytics/vince/internal/cmd"
	"github.com/vinceanalytics/vince/internal/logger"
)

func main() {
	cmd := cmd.App()
	err := cmd.Run(context.Background(), os.Args)
	if err != nil {
		logger.Fail("Exited process", "err", err)
	}
}
