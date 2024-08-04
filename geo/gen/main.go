package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"time"
)

const dateFormat = "2006-01-02T15:04:05-0700"
const urlFmt = "https://download.db-ip.com/free/dbip-country-lite-%s.mmdb.gz"

func main() {
	now := time.Now()
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
		f, err := os.Create("country.gz")
		if err != nil {
			log.Fatal(err)
		}
		f.ReadFrom(res.Body)
		f.Close()
	} else {
		b, _ := httputil.DumpResponse(res, true)
		log.Fatalf("Unable to download and save the database \n %s", string(b))
	}
}
