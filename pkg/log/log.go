package log

import (
	"os"

	"github.com/rs/zerolog"
)

var Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
}

func Get() *zerolog.Logger {
	return &Logger
}
