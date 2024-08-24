package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/vinceanalytics/vince/internal/version"
	"golang.org/x/mod/semver"
)

func main() {
	flag.Parse()
	v := load()
	switch flag.Arg(0) {
	case "major":
		v.major++
	case "minor":
		v.minor++
	}
	os.WriteFile("internal/version/VERSION", []byte(v.String()), 0600)
}

type Version struct {
	major, minor int
	ts           time.Time
}

func (v *Version) String() string {
	return fmt.Sprintf("v%d.%d+%s", v.major, v.minor, v.ts.Format(time.DateOnly))
}

func load() *Version {
	m := semver.MajorMinor(version.VERSION)
	a, b, _ := strings.Cut(m, ".")
	var v Version
	v.major, _ = strconv.Atoi(a[1:])
	v.minor, _ = strconv.Atoi(b)
	v.ts = time.Now().UTC()
	return &v
}
