package control

import (
	"runtime/debug"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

const (
	APP       = "app.kubernetes.io/name"
	COMPONENT = "app.kubernetes.io/component"
	VERSION   = "app.kubernetes.io/version"
	PART      = "app.kubernetes.io/part-of"
	MANAGED   = "app.kubernetes.io/managed-by"
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
		COMPONENT: "server",
		VERSION:   version,
		MANAGED:   "v8s",
		PART:      "vinceanalytics",
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
