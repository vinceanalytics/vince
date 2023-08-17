package must

import (
	"fmt"
	"os"

	"log/slog"
)

func Must[T any](r T, err error) func(msg string, args ...any) T {
	return func(msg string, args ...any) T {
		if err != nil {
			slog.Error(msg, append([]any{slog.String("err", err.Error())}, args...)...)
			os.Exit(1)
		}
		return r
	}
}

func One(err error) func(msg string, args ...any) {
	return func(msg string, args ...any) {
		if err != nil {
			slog.Error(msg, append([]any{slog.String("err", err.Error())}, args...)...)
			os.Exit(1)
		}
	}
}

func Assert(ok bool) func(msg ...any) {
	return func(msg ...any) {
		if !ok {
			slog.Error(fmt.Sprint(msg...))
			os.Exit(1)
		}
	}
}

func AssertFMT(ok bool) func(msg string, a ...any) {
	return func(msg string, a ...any) {
		if !ok {
			slog.Error(fmt.Sprintf(msg, a...))
			os.Exit(1)
		}
	}
}
