package main

import (
	"os"

	"github.com/gernest/vince/tools"
)

func main() {
	if os.Getenv("SITE") != "" {
		root := tools.RootVince()
		tools.ExecPlainWithWorkingPath(
			root,
			"npm", "run", "docs:build",
		)
		tools.ExecPlainWithWorkingPath(
			root,
			"npm", "run", "blog:build",
		)
	}
}
