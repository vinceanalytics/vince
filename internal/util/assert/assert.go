package assert

import (
	"log"
	"log/slog"
	"os"
)

func True(v bool) {
	if !v {
		log.Fatal("assertion failure")
	}
}

func Nil(err error, msg ...string) {
	if err != nil {
		if len(msg) > 0 {
			slog.Error(msg[0], "err", err)
		} else {
			slog.Error(err.Error())
		}
		os.Exit(1)
	}
}
