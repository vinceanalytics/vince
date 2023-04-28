package main

import (
	_ "capnproto.org/go/capnp/v3"
	"github.com/gernest/vince/tools"
)

func main() {
	println("### Generating capnp for store")
	root := tools.ModuleRoot("capnproto.org/go/capnp/v3")
	println(">>> using codegen: ", root)
	tools.ExecPlain("capnp",
		"--verbose",
		"compile", "-I"+root+"/std",
		"-ogo", "store.capnp",
	)
}
