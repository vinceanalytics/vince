package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/apache/arrow/go/v15/arrow/ipc"
	"github.com/apache/arrow/go/v15/arrow/memory"
	"github.com/blevesearch/vellum"
	"github.com/vinceanalytics/staples/staples/staples"
	"gopkg.in/yaml.v2"
)

type Ref struct {
	Name     string
	Category string
	Domains  []string
}

type Min struct {
	Name string
}

const srcFileURL = "https://raw.githubusercontent.com/snowplow-referer-parser/referer-parser/master/resources/referers.yml"

type refererData map[string]map[string]map[string][]string

func (r refererData) Ref() (o []*Ref) {
	for g, m := range r {
		for ref, domains := range m {
			o = append(o, &Ref{
				Name:     ref,
				Category: g,
				Domains:  domains["domains"],
			})
		}
	}
	sort.Slice(o, func(i, j int) bool {
		x := strings.Compare(o[i].Category, o[j].Category)
		if x == 0 {
			return strings.Compare(o[i].Name, o[j].Name) == -1
		}
		return x == -1
	})
	return o
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
	type Domain struct {
		Name  []byte
		Index uint64
	}

	var all []*Domain
	a := staples.NewArrow[Min](memory.DefaultAllocator)
	var m Min
	for i, r := range data.Ref() {
		m.Name = r.Name
		a.Append(&m)
		for _, d := range r.Domains {
			all = append(all, &Domain{
				Name:  []byte(d),
				Index: uint64(i),
			})
		}
	}
	r := a.NewRecord()
	var b bytes.Buffer
	w := ipc.NewWriter(&b, ipc.WithSchema(r.Schema()), ipc.WithZstd())
	err = w.Write(r)
	if err != nil {
		log.Fatal(err)
	}
	err = w.Close()
	if err != nil {
		log.Fatal(err)
	}
	os.WriteFile("refs.arrow", b.Bytes(), 0600)
	b.Reset()
	fs, err := vellum.New(&b, nil)
	if err != nil {
		log.Fatal(err)
	}
	sort.Slice(all, func(i, j int) bool {
		return bytes.Compare(all[i].Name, all[j].Name) == -1
	})
	for _, v := range all {
		err = fs.Insert(v.Name, v.Index)
		if err != nil {
			log.Fatal(err)
		}
	}
	err = fs.Close()
	if err != nil {
		log.Fatal(err)
	}
	os.WriteFile("refs.fst", b.Bytes(), 0600)
}
