package main

import (
	"os"

	"github.com/gernest/vince/tools"
)

func main() {
	os.RemoveAll("repo")
	os.MkdirAll("repo", 0755)
	version := os.Getenv("VERSION")
	if version != "" {
		tools.ExecPlain(
			"helm", "package", ".", "-d", "repo", "--app-version", version,
		)
	} else {
		tools.ExecPlain(
			"helm", "package", ".", "-d", "repo",
		)
	}
}
