package main

import "github.com/vinceanalytics/vince/tools"

func main() {
	tools.ExecPlain("go", "install", "-tags", "sqlite_json")
}
