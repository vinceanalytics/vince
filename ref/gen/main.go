package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sort"

	"github.com/blevesearch/vellum"
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
	var o bytes.Buffer
	ls, all := data.Ref()
	fmt.Fprintf(&o, `
	package ref
	var all_referrals =%#v
	`, ls)

	result, err := format.Source(o.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	os.WriteFile("generated.go", result, 0600)
	o.Reset()
	fs, err := vellum.New(&o, nil)
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
	os.WriteFile("refs.fst", o.Bytes(), 0600)

	os.Mkdir("favicon", 0755)
	seen := map[uint64]struct{}{}
	for _, d := range all {
		if _, ok := seen[d.Index]; ok {
			continue
		}
		data := get(string(d.Name))
		if len(data) == 0 {
			continue
		}
		seen[d.Index] = struct{}{}
		os.WriteFile(filepath.Join("favicon", ls[d.Index]), data, 0600)
	}
}

var klient = &http.Client{}

func get(domain string) []byte {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("https://icons.duckduckgo.com/ip3/%s.ico", domain), nil)
	res, err := klient.Do(req)
	if err != nil {
		fmt.Println(">", domain)
		return []byte{}
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusOK {
		var b bytes.Buffer
		io.Copy(&b, res.Body)
		if !bytes.Contains(b.Bytes(), []byte{137, 80, 78, 71, 13, 10, 26, 10}) {
			return b.Bytes()
		}
	}
	fmt.Println(">", domain)
	return []byte{}
}
