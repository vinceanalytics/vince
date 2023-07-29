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
