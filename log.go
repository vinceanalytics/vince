package vince

import (
	"os"

	"github.com/rs/zerolog"
)

var xlg = zerolog.New(os.Stderr).With().Timestamp().Logger()

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func setDebug() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}
