package main

import (
	"os"
	"os/exec"

	"github.com/vinceanalytics/vince/tools"
)

func main() {
	if os.Getenv("BUILD") != "" {
		tools.ExecPlainWith(
			func(c *exec.Cmd) {
				c.Dir = tools.RootVince()
			},
			"goreleaser", "release", "--clean", "--timeout", "60m",
		)
	}
}
