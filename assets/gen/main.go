package main

import (
	"path/filepath"

	"github.com/gernest/vince/tools"
)

func main() {
	println("### Generating  assets ###")
	root := tools.RootVince()
	println(">>> root:", root)
	tools.ExecPlain(
		"sass", "-scompressed", "--no-source-map",
		filepath.Join(root, "assets/ui/scss/main.scss"),
		filepath.Join(root, "assets/css/app.css"),
	)
}
