package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

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

	path := "k8s/Chart.yaml"
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	data = bytes.ReplaceAll(data, []byte(version.VERSION), []byte(v.String()))
	err = os.WriteFile(path, data, 0600)
	if err != nil {
		log.Fatal(err)
	}
}

type Version struct {
	major, minor, patch int
}

func (v *Version) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.major, v.minor, v.patch)
}

func load() *Version {
	m, _, _ := strings.Cut(version.VERSION, "+")
	p := strings.Split(m, ".")
	var v Version
	v.major, _ = strconv.Atoi(p[0][1:])
	v.minor, _ = strconv.Atoi(p[1])
	v.patch, _ = strconv.Atoi(p[2])
	return &v
}
