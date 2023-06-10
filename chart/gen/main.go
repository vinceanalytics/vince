package main

import (
	"github.com/vinceanalytics/vince/tools"
)

func main() {
	tools.ExecPlain(
		"helm", "package", ".", "-d", "charts",
	)
	o := tools.ExecCollect("helm", "template", ".")
	tools.WriteFile("charts/v8s.yaml", []byte(o))
}
