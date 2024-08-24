package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/vinceanalytics/vince/internal/version"
)

func main() {
	flag.Parse()
	v := load()
	switch flag.Arg(0) {
	case "major":
		v.major++
	case "minor":
		v.minor++
	case "patch":
		v.patch++
	}
	os.WriteFile("internal/version/VERSION", []byte(v.String()), 0600)
}

type Version struct {
	major, minor, patch int
	ts                  time.Time
}

func (v *Version) String() string {
	return fmt.Sprintf("v%d.%d.%d+%s", v.major, v.minor, v.patch, v.ts.Format(version.TimeFormat))
}

func load() *Version {
	m, _, _ := strings.Cut(version.VERSION, "+")
	p := strings.Split(m, ".")
	var v Version
	v.major, _ = strconv.Atoi(p[0][1:])
	v.minor, _ = strconv.Atoi(p[1])
	v.patch, _ = strconv.Atoi(p[2])
	v.ts = time.Now().UTC()
	return &v
}
