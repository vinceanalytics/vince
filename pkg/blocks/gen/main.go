package main

import "github.com/vinceanalytics/vince/tools"

func main() {
	tools.ExecPlain("protoc",
		"-I=.", "--go_out=paths=source_relative:.", "blocks.proto",
	)
}
