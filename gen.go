package vince

//go:generate protoc -I=. --go_out=paths=source_relative:. events.proto
//go:generate go run referrer/make_referrer.go
//go:generate go run ua/bot/make_bot.go
