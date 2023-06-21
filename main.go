package main

import (
	"github.com/vinceanalytics/vince/internal/cmd/vince"
	"github.com/vinceanalytics/vince/pkg/run"
)

//go:generate go run tools/install/main.go

func main() {
	run.Main(vince.App())
}
