package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"math/rand"
	"sync"

	"github.com/klauspost/compress/zstd"
)

//go:embed ip_list.zstd
var ipListFile []byte

var ipList []string
var ipOnce sync.Once

func ip() string {

	ipOnce.Do(func() {
		r, err := zstd.NewReader(bytes.NewReader(ipListFile))
		if err != nil {
			panic(err.Error())
		}
		defer r.Close()
		err = json.NewDecoder(r).Decode(&ipList)
		if err != nil {
			panic(err.Error())
		}
	})
	n := rand.Intn(len(ipList))
	return ipList[n]
}
