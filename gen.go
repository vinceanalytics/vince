package vince

//go:generate protoc -I=. --go_out=paths=source_relative:. events.proto
//go:generate go run ua/bot/make_bot.go
//go:generate go run ua/device/make_device.go
//go:generate go run ua/client/make_client.go
//go:generate go run ua/os/make_os.go
//go:generate go run ua/vendor_fragment/make_vendor.go
//go:generate go run schema/make_schema.go
//go:generate go run referrer/make_referrer.go
