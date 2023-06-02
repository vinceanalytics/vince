package main

import (
	"github.com/vinceanalytics/vince/cmd/app/vince"
	"github.com/vinceanalytics/vince/cmd/run"
)

func main() {
	run.Main(vince.App())
}
