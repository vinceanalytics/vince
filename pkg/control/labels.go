package control

import (
	"runtime/debug"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

const (
	APP       = "app"
	COMPONENT = "component"
	VERSION   = "version"
)

var version string

func init() {
	build, _ := debug.ReadBuildInfo()
	version = build.Main.Version
}

// all resources managed by this controller will be created with these labels. We
// tie resources to the specific v8s version.
func baseLabels() map[string]string {
	return map[string]string{
		APP:       "vince",
		COMPONENT: "v8s",
		VERSION:   version,
	}
}

func baseSelector() labels.Selector {
	b := baseLabels()
	s := labels.NewSelector()
	for k, v := range b {
		r, err := labels.NewRequirement(k, selection.Equals, []string{v})
		if err != nil {
			panic("failed to create requirement " + err.Error())
		}
		s = s.Add(*r)
	}
	return s
}
