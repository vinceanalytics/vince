package main

import (
	"bytes"
	"flag"
	"io"
	"net/http"
	"os"
	"slices"

	"log"

	"github.com/cespare/xxhash/v2"
	"github.com/vinceanalytics/vince/fb"
	"github.com/vinceanalytics/vince/internal/roaring"
	"gopkg.in/yaml.v2"
)

type Ref struct {
	Name     string
	Index    int
	Category string
	Domains  []string
}

type Min struct {
	Name string
}

const srcFileURL = "https://raw.githubusercontent.com/snowplow-referer-parser/referer-parser/master/resources/referers.yml"

type refererData map[string]map[string]map[string][]string

func (r refererData) Ref() (b []string, domains []*Domain) {
	re := map[string]int{}
	var o []*Ref
	var size int
	for g, m := range r {
		for ref, domains := range m {
			re[ref] = 0
			r := &Ref{
				Name:     ref,
				Category: g,
				Domains:  domains["domains"],
			}
			o = append(o, r)
			size += len(domains)
		}
	}
	b = make([]string, 0, len(re))
	for k := range re {
		b = append(b, k)
	}
	slices.Sort(b)
	for i := range b {
		re[b[i]] = i
	}
	domains = make([]*Domain, 0, size)
	for _, r := range o {
		idx := uint64(re[r.Name])
		for _, n := range r.Domains {
			domains = append(domains, &Domain{
				Name:  []byte(n),
				Index: idx,
			})
		}
	}

	return
}

type Domain struct {
	Name  []byte
	Index uint64
}

func main() {
	flag.Parse()
	res, err := http.Get(srcFileURL)
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		log.Fatal(res.Status)
	}

	bs, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	var data refererData
	err = yaml.NewDecoder(bytes.NewReader(bs)).Decode(&data)
	if err != nil {
		log.Fatal(err)
	}
	ls, all := data.Ref()
	bsi := roaring.NewDefaultBSI()

	domains := make([]string, 0, len(all))

	for _, v := range all {
		bsi.SetValue(
			xxhash.Sum64(v.Name), int64(v.Index),
		)

	}
	out := fb.SerializeRef(ls, bsi.ToBuffer())
	os.WriteFile("refs.fbs.bin", out, 0600)
	slices.Sort(domains)
}
