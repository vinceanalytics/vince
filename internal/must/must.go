package must

import "github.com/vinceanalytics/vince/pkg/log"

func Must[T any](r T, err error) T {
	if err != nil {
		log.Get().Fatal().Err(err).Msg("failed a must condition")
	}
	return r
}
