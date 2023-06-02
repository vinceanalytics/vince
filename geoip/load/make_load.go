package main

import (
	"encoding/json"
	"log"
	"os"
	"sort"

	"github.com/klauspost/compress/zstd"
	"github.com/vinceanalytics/vince/geoip"
)

var ps = map[string]struct{}{}

func main() {
	r := geoip.Reader()
	n := r.Networks()
	for n.Next() {
		var v interface{}
		ip, err := n.Network(&v)
		if err != nil {
			log.Fatal(err)
		}
		ps[ip.IP.String()] = struct{}{}
	}
	var ls []string
	for k := range ps {
		ls = append(ls, k)
	}
	sort.Strings(ls)
	f, err := os.Create("cmd/load/ip_list.zstd")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	w, err := zstd.NewWriter(f, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
	if err != nil {
		log.Fatal(err)
	}
	defer w.Close()
	err = json.NewEncoder(w).Encode(ls)
	if err != nil {
		log.Fatal(err)
	}
}
