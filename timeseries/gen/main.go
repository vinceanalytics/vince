package main

import "github.com/gernest/vince/tools"

func main() {
	println("### Generating proto for timeseries ###")
	tools.ExecPlain("protoc",
		"-I=.", "--go_out=paths=source_relative:.", "event.proto",
	)
}
