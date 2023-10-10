package main

import (
	"github.com/vinceanalytics/vince/internal/cmd/vince"
	"github.com/vinceanalytics/vince/internal/run"
)

func main() {
	run.Main(vince.App())
}
