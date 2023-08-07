package main

import "github.com/vinceanalytics/vince/tools"

func main() {
	tools.ExecPlain("go", "generate", "./ui")
	tools.ExecPlain("go", "run", "main.go")
}
