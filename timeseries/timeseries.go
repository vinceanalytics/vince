package timeseries

import "github.com/gernest/vince/system"

//go:generate protoc -I=. --go_out=paths=source_relative:. event.proto

var (
	// Distribution of save duration for each accumulated buffer
	saveDuration = system.Get("mike_save_duration")
)
