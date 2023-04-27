package main

import (
	"fmt"
	"runtime/debug"

	_ "capnproto.org/go/capnp/v3"
	"github.com/gernest/vince/tools"
)

func main() {
	println("### Generating capnp for store")
	build, _ := debug.ReadBuildInfo()
	CODEGEN_PKG := fmt.Sprintf("%s/pkg/mod/%s@%s",
		tools.ExecCollect("go", "env", "GOPATH"),
		build.Deps[0].Path, build.Deps[0].Version,
	)
	println(">>> using codegen: ", CODEGEN_PKG)
	tools.ExecPlain("capnp",
		"--verbose",
		"compile", "-I"+CODEGEN_PKG+"/std",
		"-ogo", "store.capnp",
	)
}
