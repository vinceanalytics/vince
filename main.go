package main

import (
	"github.com/vinceanalytics/vince/internal/cmd/vince"
	"github.com/vinceanalytics/vince/pkg/run"
)

func main() {
	run.Main(vince.App())
}
