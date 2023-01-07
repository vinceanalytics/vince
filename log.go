package vince

import (
	"os"

	"github.com/rs/zerolog"
)

var xlg = zerolog.New(os.Stderr).With().Timestamp().Logger()

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
}
