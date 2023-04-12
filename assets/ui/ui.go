package ui

import (
	"embed"
)

//go:generate sass --no-source-map scss/main.scss app/css/app.css
//go:generate go run gen/main.go app ../../assets/

//go:embed app
var UI embed.FS
