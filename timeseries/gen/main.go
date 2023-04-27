package main

import "github.com/gernest/vince/tools"

func main() {
	println("### Generating proto for timeseries ###")
	tools.Proto("event.proto")
}
