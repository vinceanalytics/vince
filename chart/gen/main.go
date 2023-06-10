package main

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/vinceanalytics/vince/tools"
)

func main() {
	tools.ExecPlain(
		"helm", "package", ".", "-d", "charts",
	)
	var b bytes.Buffer
	fs, err := os.ReadDir("crds")
	if err != nil {
		tools.Exit(err.Error())
	}
	for _, f := range fs {
		b.Write(tools.ReadFile(filepath.Join("crds", f.Name())))
	}
	b.WriteString(tools.ExecCollect("helm", "template", "."))

	dir := filepath.Join(tools.RootVince(), "website/docs/.vitepress/dist/kustomize")
	tools.MkDir(dir)
	tools.WriteFile(filepath.Join(dir, "v8s.yaml"), b.Bytes())
}
