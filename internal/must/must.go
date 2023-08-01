package must

import (
	"fmt"

	"github.com/vinceanalytics/vince/pkg/log"
)

func Must[T any](r T, err error) func(msg ...any) T {
	return func(msg ...any) T {
		if err != nil {
			log.Get().Fatal().Err(err).Msg(fmt.Sprint(msg...))
		}
		return r
	}
}

func One(err error) func(msg ...any) {
	return func(msg ...any) {
		if err != nil {
			log.Get().Fatal().Err(err).Msg(fmt.Sprint(msg...))
		}
	}
}

func Assert(ok bool) func(msg ...any) {
	return func(msg ...any) {
		if !ok {
			log.Get().Fatal().Msg(fmt.Sprint(msg...))
		}
	}
}

func AssertFMT(ok bool) func(msg string, a ...any) {
	return func(msg string, a ...any) {
		if !ok {
			log.Get().Fatal().Msgf(msg, a...)
		}
	}
}
