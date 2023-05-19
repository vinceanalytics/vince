package main

import (
	"os"
	"path/filepath"

	"github.com/gernest/vince/tools"
)

func main() {
	dir, err := os.ReadDir(".")
	if err != nil {
		tools.Exit(err.Error())
	}
	for _, f := range dir {
		if f.IsDir() {
			continue
		}
		if filepath.Ext(f.Name()) == ".tgz" {
			tools.Remove(f.Name())
		}
	}
	tools.ExecPlain(
		"helm", "package", ".",
	)
}
