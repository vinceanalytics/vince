package main

import (
	"github.com/vinceanalytics/vince/cmd/app/v8s"
	"github.com/vinceanalytics/vince/cmd/run"
)

func main() {
	run.Main(v8s.App())
}
