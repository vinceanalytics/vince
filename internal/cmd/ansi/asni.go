package ansi

import (
	"fmt"
	"os"
)

const esc = "\033["

func Green(s string) string {
	return fmt.Sprintf("%s32m%v%s", esc, s, resetCode)
}

func Black(s string) string {
	return fmt.Sprintf("%s30m%v%s", esc, s, resetCode)
}

func Red(s string) string {
	return fmt.Sprintf("%s31m%v%s", esc, s, resetCode)
}

const (
	Check     = "✓"
	Selection = "→"
	Zap       = "⚡"
	X         = "✗"
)

var resetCode = fmt.Sprintf("%s%dm", esc, 0)

func Ok(msg string, args ...any) {
	fmt.Fprintf(os.Stdout, " %s %s\n", Green(Check), fmt.Sprintf(msg, args...))
}

func Step(msg string, args ...any) {
	fmt.Fprintf(os.Stdout, "%s %s\n", Black(Selection), fmt.Sprintf(msg, args...))
}

func Err(msg string, args ...any) {
	fmt.Fprintf(os.Stdout, "%s %s\n", Red(X), fmt.Sprintf(msg, args...))
}

func ERROR(err error) error {
	if err == nil {
		return nil
	}
	Err(err.Error())
	return err
}

func Suggestion(ls ...string) {
	fmt.Fprintln(os.Stdout, "try:")
	for _, k := range ls {
		Step(k)
	}
}
