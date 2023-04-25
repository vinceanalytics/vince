package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

func main() {
	var all []Artifact
	flag.Parse()
	b, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(b, &all)
	var o bytes.Buffer
	fmt.Fprintf(&o, "| filename | signature | size |\n")
	fmt.Fprintf(&o, "| ---- | ---- | ---- |\n")
	for _, a := range all {
		if a.Type != "Archive" {
			continue
		}
		stat, err := os.Stat(a.Path)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(&o, "| [%s]() | [minisig](%s.minisig)  | `%s` |\n",
			a.Name, a.Name, size(int(stat.Size())),
		)
	}
	os.Stdout.WriteString(o.String())
}

func size(n int) string {
	if n < (1 << 20) {
		return strconv.Itoa(n/(1<<10)) + "kb"
	}
	if n < (1 << 30) {
		return strconv.Itoa(n/(1<<20)) + "mb"
	}
	return strconv.Itoa(n)
}

type Artifact struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Path string `json:"path"`
}
