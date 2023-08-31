package main

import "github.com/vinceanalytics/vince/tools"

func main() {
	tools.ExecPlainWithWorkingPath(
		"./v1",
		"protoc",
		"-I=.",
		"--go_out=paths=source_relative:.",
		"--go-grpc_out=paths=source_relative:.",
		"api.proto",
	)
}
