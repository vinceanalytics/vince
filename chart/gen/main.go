package main

import (
	"path/filepath"

	"github.com/vinceanalytics/vince/tools"
)

func main() {
	tools.ExecPlain(
		"helm", "package", ".", "-d", "charts",
	)
	o := tools.ExecCollect("helm", "template", ".")
	dir := filepath.Join(tools.RootVince(), "website/docs/.vitepress/dist/kustomize")
	tools.MkDir(dir)
	tools.WriteFile(filepath.Join(dir, "v8s.yaml"), []byte(o))
}
