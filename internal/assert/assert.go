package assert

import (
	"log/slog"
	"os"
)

func Assert(cond bool, msg string, args ...any) {
	if !cond {
		slog.Error(msg, args...)
		os.Exit(1)
	}
}
