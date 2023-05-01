package main

import (
	"github.com/gernest/vince/cmd/app/v8s"
	"github.com/gernest/vince/cmd/run"
)

func main() {
	run.Main(v8s.App())
}
