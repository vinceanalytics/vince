package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"os"
	"sort"

	"github.com/klauspost/compress/zstd"
	"github.com/oschwald/maxminddb-golang"
	"github.com/vinceanalytics/vince/internal/geoip"
	"github.com/vinceanalytics/vince/internal/must"
)

var ps = map[string]struct{}{}

func main() {
	g := must.Must(gzip.NewReader(bytes.NewReader(geoip.CityData)))(
		"failed to read embedded mmdb data file gzip data expected ",
	)
	b := must.Must(io.ReadAll(g))("failed reading compressed mmdb")
	r := must.Must(maxminddb.FromBytes(b))(
		"failed opening mmdb",
	)
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
	f, err := os.Create("ip_list.zstd")
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
