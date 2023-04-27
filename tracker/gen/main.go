package main

import "github.com/gernest/vince/tools"

func main() {
	println("### Generating tracking script ###")
	tools.ExecPlain("node", "compile.js")
}
