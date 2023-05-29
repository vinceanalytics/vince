package main

import (
	"os"
	"path/filepath"

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
		from := filepath.Join(root, "blog/.vitepress/dist/") + "/"
		to := filepath.Join(root, "docs/.vitepress/dist/blog/")
		tools.CopyDir(from, to)
	}
}
