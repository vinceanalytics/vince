package timeseries

//go:generate protoc -I=. --go_out=paths=source_relative:. event.proto
//go:generate go run golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment -fix  .
