package main

import (
	"path/filepath"

	"github.com/gernest/vince/tools"
)

var root string

func main() {
	println("### Generating  assets ###")
	var err error
	root, err = filepath.Abs("../")
	if err != nil {
		tools.Exit("failed to resolve root")
	}
	println(">>> root:", root)
	sass()
}

func sass() {
	tools.ExecPlain(
		"sass", "-scompressed", "--no-source-map",
		filepath.Join(root, "assets/ui/scss/main.scss"),
		filepath.Join(root, "assets/css/app.css"),
	)
}
