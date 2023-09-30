package main

import (
	"github.com/vinceanalytics/vince/internal/cmd/vince"
	"github.com/vinceanalytics/vince/internal/run"
)

//go:generate gotip generate ./ui
//go:generate gotip run main.go serve .vince

func main() {
	run.Main(vince.App())
}
