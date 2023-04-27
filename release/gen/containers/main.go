package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gernest/vince/tools"
)

const (
	labelCreated  = "org.opencontainers.image.created=%q"
	labelRevision = "org.opencontainers.image.revision=%q"
	labelTitle    = "org.opencontainers.image.title=%q"
)

var root string

func main() {
	println("### Building container image ###")
	var err error
	root, err = filepath.Abs("../")
	if err != nil {
		tools.Exit("failed to resolve project root", err.Error())
	}
	println(">>> root: ", root)
	make()
}

func make() {
	meta, a := tools.Release(root)
	dockerFile, err := os.ReadFile(filepath.Join(root, "Dockerfile"))
	if err != nil {
		tools.Exit("failed to read docker  file", err.Error())
	}
	for _, v := range a {
		if v.Type != "Binary" || v.Os != "linux" {
			continue
		}
		// copy docker file to the context
		tools.WriteFile(filepath.Join(filepath.Dir(v.Path), "Dockerfile"), dockerFile)

		tools.ExecPlain(
			"docker",
			"--log-level", "error",
			"build",
			"--platform", v.Os+"/"+v.Arch,
			"--label", fmt.Sprintf(labelCreated, meta.Date.UTC().Format(time.RFC3339)),
			"--label", fmt.Sprintf(labelRevision, meta.Commit),
			"--label", fmt.Sprintf(labelTitle, meta.Image(v.Arch)),
			"--tag", meta.Image(v.Arch),
			filepath.Dir(v.Path),
		)
		println(">>> tag:", meta.Image(v.Arch))
	}
	tools.ExecPlain("docker", "tag", meta.Image(runtime.GOARCH), meta.Latest())
	println(">>> tag:", meta.Latest())
}
