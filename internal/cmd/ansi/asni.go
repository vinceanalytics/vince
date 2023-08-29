package ansi

import (
	"fmt"
	"os"
	"text/tabwriter"
)

const (
	ok   = "✓"
	step = "→"
	zap  = "⚡"
	x    = "✗"
)

type W struct {
	*tabwriter.Writer
}

func New() *W {
	return &W{
		Writer: tabwriter.NewWriter(
			os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight,
		),
	}
}

func (w *W) Step(msg string) *W {
	fmt.Fprintf(w, "%s \t%s\n", step, msg)
	return w
}

func (w *W) Ok(msg string, a ...any) *W {
	fmt.Fprintf(w, "%s \t%s\n", ok, fmt.Sprintf(msg, a...))
	return w
}

func (w *W) Err(msg string, a ...any) *W {
	fmt.Fprintf(w, "%s \t%s\n", x, fmt.Sprintf(msg, a...))
	return w
}

func (w *W) KV(key, value string, args ...any) *W {
	fmt.Fprintf(w, "%s \t%s\n", key, fmt.Sprintf(value, args...))
	return w
}

func (w *W) Suggest(msg ...string) *W {
	fmt.Fprintln(w, "suggestion \t ")
	for _, m := range msg {
		fmt.Fprintf(w, " \t%s\n", m)
	}
	return w
}

func (w *W) Complete(err error) error {
	if err != nil {
		w.Err(err.Error())
		w.Flush()
		return err
	}
	return w.Flush()
}

func (w *W) Exit() {
	w.Flush()
	os.Exit(1)
}
