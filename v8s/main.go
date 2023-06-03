package main

import (
	"github.com/vinceanalytics/vince/pkg/run"
	"github.com/vinceanalytics/vince/v8s/cmd/v8s"
)

func main() {
	run.Main(v8s.App())
}
