package version

import (
	_ "embed"
	"fmt"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

//go:embed VERSION
var VERSION string

const TimeFormat = "20060102"

func Build() time.Time {
	b := semver.Build(VERSION)
	fmt.Println(b)
	b = strings.TrimPrefix(b, "+")
	if b == "" {
		return time.Time{}
	}
	ts, _ := time.Parse(TimeFormat, b)
	return ts
}
