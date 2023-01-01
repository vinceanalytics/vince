package vince

//go:generate protoc -I=. --go_out=paths=source_relative:. events.proto
//go:generate go run referrer/make_referrer.go
