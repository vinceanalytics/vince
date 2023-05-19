package main

import (
	"flag"
	"os"

	"github.com/gernest/vince/tools"
)

func main() {
	flag.Parse()
	os.RemoveAll("repo")
	os.MkdirAll("repo", 0755)
	tools.ExecPlain(
		"helm", "package", ".",
	)
}
