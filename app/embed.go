package app

import (
	"embed"
)

//go:embed public
var Public embed.FS

//go:embed images
var Images embed.FS
