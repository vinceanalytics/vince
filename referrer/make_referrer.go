package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"go/format"
	"log"
	"net/url"
	"os"
	"sort"
	"strings"
)

//go:embed referrer.json
var refererJSON string

var index = map[string]bool{}

func main() {
	m := make(map[string]interface{})
	json.Unmarshal([]byte(refererJSON), &m)
	var maxLen int
	var minLen = 6

	var b bytes.Buffer
	var hosts []*Medium
	for mTyp, mData := range m {
		for refName, refConfig := range mData.(map[string]interface{}) {
			var params []string
			if p, ok := refConfig.(map[string]interface{})["parameters"]; ok {
				for _, v := range p.([]interface{}) {
					params = append(params, v.(string))
				}
			}
			for _, domain := range refConfig.(map[string]interface{})["domains"].([]interface{}) {
				med := &Medium{
					Type:       mTyp,
					Name:       refName,
					Parameters: params,
				}
				dm := domain.(string)
				u, _ := url.Parse("http://" + dm)
				host := strings.TrimPrefix(u.Host, "www.")
				parts := strings.Split(host, ".")
				sort.Sort(sort.Reverse(StringSlice(parts)))
				if len(parts) > int(maxLen) {
					maxLen = len(parts)
				}
				host = strings.Join(parts, ".")
				if len(parts) < minLen {
					minLen = len(parts)
				}
				if index[host] {
					continue
				}
				hosts = append(hosts, med)
				med.Host = host
				index[host] = true
			}
		}
	}
	fmt.Fprintln(&b, "// DO NOT EDIT Code generated by referrer/make_referrer.go")
	fmt.Fprintln(&b, " package vince")
	fmt.Fprintf(&b, " const minReferrerSize=%d\n", minLen)
	fmt.Fprintf(&b, " const maxReferrerSize=%d\n", maxLen)
	fmt.Fprintln(&b, `
	type Medium struct {
		Type       string
		Name       string
	}
	`)
	fmt.Fprintln(&b, "var refList=map[string]*Medium{")
	sort.Sort(MedSLice(hosts))
	for _, m := range hosts {
		fmt.Fprintf(&b, "%q:{Type:%q,Name:%q},\n", m.Host, m.Type, m.Name)
	}
	fmt.Fprintln(&b, "}")

	r, err := format.Source(b.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	os.WriteFile("referrer_data.go", r, 0600)
}

type StringSlice []string

func (x StringSlice) Len() int           { return len(x) }
func (x StringSlice) Less(i, j int) bool { return i < j }
func (x StringSlice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type Medium struct {
	Host       string
	Type       string
	Name       string
	Parameters []string
}

type MedSLice []*Medium

func (x MedSLice) Len() int { return len(x) }
func (x MedSLice) Less(i, j int) bool {
	a := x[i]
	b := x[j]
	return (a.Host + a.Type + a.Name) < (b.Host + b.Type + b.Name)
}
func (x MedSLice) Swap(i, j int) { x[i], x[j] = x[j], x[i] }
