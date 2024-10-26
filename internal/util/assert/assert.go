package assert

import (
	"errors"
	"log"
	"log/slog"
	"os"

	"github.com/cockroachdb/pebble"
)

func True(v bool) {
	if !v {
		log.Fatal("assertion failure")
	}
}

func Nil(err error, msg ...string) {
	if err != nil && !errors.Is(err, pebble.ErrNotFound) {
		if len(msg) > 0 {
			slog.Error(msg[0], "err", err)
		} else {
			slog.Error(err.Error())
		}
		os.Exit(1)
	}
}
