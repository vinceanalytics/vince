package main

import (
	"github.com/vinceanalytics/vince/internal/cmd/vince"
	"github.com/vinceanalytics/vince/internal/run"
)

//go:generate go generate ./ui
//go:generate go run main.go serve .vince

func main() {
	run.Main(vince.App())
}
