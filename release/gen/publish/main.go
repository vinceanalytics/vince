package main

import (
	"os"
	"os/exec"

	"github.com/vinceanalytics/vince/tools"
)

func main() {
	root := tools.RootVince()
	if os.Getenv("SITE") != "" {
		tools.ExecPlainWith(func(c *exec.Cmd) {
			c.Dir = root
			c.Env = append(c.Env, "SITE=true")
		},
			"go", "generate", "./website",
		)
	}
}
