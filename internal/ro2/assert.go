package ro2

import (
	"log/slog"
	"os"
)

func assert(cond bool, msg string, args ...any) {
	if !cond {
		slog.Error(msg, args...)
		os.Exit(1)
	}
}
