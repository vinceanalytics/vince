package tracker

import "embed"

//go:generate go run gen/main.go js

//go:embed js
var JS embed.FS
