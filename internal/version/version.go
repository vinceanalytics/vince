package version

import (
	_ "embed"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

//go:embed VERSION
var VERSION string

func Build() time.Time {
	b := semver.Build(VERSION)
	b = strings.TrimPrefix(b, "+")
	if b == "" {
		return time.Time{}
	}
	ts, _ := time.Parse(time.DateOnly, b)
	return ts
}
