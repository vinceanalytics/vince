package main

import (
	"bytes"
	"compress/gzip"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"time"

	"github.com/vinceanalytics/vince/internal/util/assert"
	"github.com/vinceanalytics/vince/internal/util/xtime"
)

const dateFormat = "2006-01-02T15:04:05-0700"
const urlFmt = "https://download.db-ip.com/free/dbip-city-lite-%s.mmdb.gz"

func main() {
	flag.Parse()
	if flag.Arg(0) == "chunk" {
		writeChunked(flag.Arg(1))
		return
	}
	now := xtime.Now()
	thisMoth := now.Format(dateFormat)[0:7]
	lastMonth := time.Date(
		now.Year(), now.Month()-1,
		0, 0, 0, 0, 0, now.Location(),
	).Format(dateFormat)[0:7]
	var b bytes.Buffer
	fmt.Fprintf(&b, urlFmt, thisMoth)
	thisMonthURL := b.String()
	b.Reset()
	fmt.Fprintf(&b, urlFmt, lastMonth)
	lastMonthURL := b.String()
	println(thisMonthURL)
	println(lastMonthURL)

	res, err := http.Get(thisMonthURL)
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode == http.StatusNotFound {
		res.Body.Close()
		log.Printf(" Got 404 for %s ,trying %s", thisMonthURL, lastMonthURL)
		res, err = http.Get(lastMonthURL)
		if err != nil {
			log.Fatal(err)
		}
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusOK {
		f, err := os.Create("city.mmdb")
		if err != nil {
			log.Fatal(err)
		}
		r, _ := gzip.NewReader(res.Body)
		f.ReadFrom(r)
		f.Close()
	} else {
		b, _ := httputil.DumpResponse(res, true)
		log.Fatalf("Unable to download and save the database \n %s", string(b))
	}
}

func writeChunked(name string) {
	os.Mkdir("data", 0755)
	f, err := os.Open(name)
	assert.Nil(err)
	defer f.Close()

	buf := make([]byte, 8<<20)
	var count int
	for {
		count++
		n, err := f.Read(buf)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				assert.Nil(err)
			}
			if n > 0 {
				writeFile(count, buf[:n])
			}
			return
		}
		writeFile(count, buf[:n])
	}
}

func writeFile(n int, data []byte) {
	path := filepath.Join("data", fmt.Sprintf("%004d.bin", n))
	assert.Nil(os.WriteFile(path, data, 0600))
}
