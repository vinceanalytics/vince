package timeseries

//go:generate protoc -I=. --go_out=paths=source_relative:. aggregate.proto
//go:generate protoc -I=. --go_out=paths=source_relative:. event.proto
