package main

import (
	"github.com/gernest/vince/cmd/app/vince"
	"github.com/gernest/vince/cmd/run"
)

func main() {
	run.Main(vince.App())
}
