package main

import (
	"path/filepath"

	"github.com/gernest/vince/pkg/plot"
	"github.com/gernest/vince/tools"
	"github.com/invopop/jsonschema"
)

func main() {
	println("### Generating json schema for plot data ###")
	root := tools.RootVince()
	j := jsonschema.Reflect(&plot.Data{})
	b, _ := j.MarshalJSON()
	tools.WriteFile(filepath.Join(root, "assets", "schema.json"), b)
}
