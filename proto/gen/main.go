package main

import "github.com/vinceanalytics/vince/tools"

func main() {
	tools.ExecPlainWithWorkingPath(
		"./v1",
		"protoc",
		"-I=.", "--go_out=paths=source_relative:.",
		"api.proto",
	)
	tools.ExecPlainWithWorkingPath(
		"./bpb",
		"protoc",
		"-I=.", "--go_out=paths=source_relative:.",
		"bpb.proto",
	)
}
