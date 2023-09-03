package main

import (
	"os"
	"path/filepath"

	"github.com/vinceanalytics/vince/internal/tools"
	"golang.org/x/mod/semver"
)

func main() {
	if os.Getenv("VERSION") != "" {
		v := tools.Version()
		update(v)
		commit(v)
		tag(v)
	}
}

func update(v string) {
	root := tools.RootVince()
	println("> using", v)
	vince(root, v)
}

func vince(root, v string) {
	println("> update vince version")
	app := tools.ReadFile(filepath.Join(root, "internal", "version", "VERSION.txt"))
	switch semver.Compare(v, string(app)) {
	case 0:
		println("> no version changes")
		os.Exit(0)
	case -1:
		tools.Exit(v, "VERSION must be greater than", string(app))
	}
	tools.WriteFile(filepath.Join(root, "internal", "version", "VERSION.txt"), []byte(v))
}

func commit(v string) {
	println("> commit", v)
	tools.ExecPlain(
		"git", "commit", "-am", "release "+v,
	)
}

func tag(v string) {
	println("> tag", v)
	tools.ExecPlain(
		"git", "tag", "-a", v, "-m", "release "+v,
	)
	tools.ExecPlain(
		"git", "push",
	)
}
