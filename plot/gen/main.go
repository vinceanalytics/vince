package main

import (
	"github.com/gernest/vince/plot"
	"github.com/gernest/vince/tools"
	"github.com/invopop/jsonschema"
)

func main() {
	println("### Generating json schema for plot data ###")
	j := jsonschema.Reflect(&plot.Data{})
	b, _ := j.MarshalJSON()
	tools.WriteFile("../assets/schema.json", b)
}
