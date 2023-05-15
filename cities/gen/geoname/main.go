package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"github.com/gernest/vince/tools"
	"github.com/klauspost/compress/zstd"
)

var client = http.Client{}

func main() {
	_, err := os.Stat(allCountriesURI)
	if err != nil {
		if os.IsNotExist(err) {
			println(">  downloading " + allCountriesURI)
			downloadCountries()
		} else {
			tools.Exit(err.Error())
		}
	} else {
		// make sure we are up to date
		println(">  using " + allCountriesURI)
	}
	processCountry()
}

type Feature struct {
	ID   string
	Name string
}

func processCountry() {
	r, err := zip.OpenReader(allCountriesURI)
	if err != nil {
		tools.Exit("failed to open zip file ", allCountriesURI, err.Error())
	}
	defer r.Close()

	o, err := r.File[0].Open()
	if err != nil {
		tools.Exit("failed to read zip file ", allCountriesURI, err.Error())
	}
	defer o.Close()

	db, err := badger.Open(badger.DefaultOptions("").
		WithInMemory(true).WithLoggingLevel(badger.ERROR))
	if err != nil {
		tools.Exit("failed to open badger db  ", err.Error())
	}
	defer db.Close()
	s := bufio.NewScanner(o)
	var key [4]byte

	sParse(s, 0, func(raw []string) bool {
		if len(raw) != 19 {
			return true
		}
		id := raw[0]
		name := raw[1]
		class := raw[6]
		if class == "P" {
			err = db.Update(func(txn *badger.Txn) error {
				i, err := strconv.Atoi(id)
				if err != nil {
					return err
				}
				binary.BigEndian.PutUint32(key[:], uint32(i))
				return txn.Set(key[:], []byte(name))
			})
			if err != nil {
				tools.Exit("failed to update  badger ", err.Error())
			}
		}
		return true
	})

	f, err := os.Create("city_geoname_db.zstd")
	if err != nil {
		tools.Exit("failed to create  file city_geoname_db.zstd  ", err.Error())
	}
	defer f.Close()
	println("   write: ", f.Name())
	e, _ := zstd.NewWriter(f, zstd.WithEncoderLevel(
		zstd.SpeedBestCompression,
	))
	_, err = db.Backup(e, 0)
	if err != nil {
		tools.Exit("failed to create  cities backup  ", err.Error())
	}
	e.Close()
}

func downloadCountries() {
	url := fmt.Sprintf("%s%s", geonamesURL, allCountriesURI)
	res, err := client.Get(url)
	if err != nil {
		tools.Exit("failed to download countries", url, err.Error())
	}
	defer res.Body.Close()
	f, err := os.Create(allCountriesURI)
	if err != nil {
		tools.Exit("failed to create  ", allCountriesURI, err.Error())
	}
	defer f.Close()
	_, err = f.ReadFrom(res.Body)
	if err != nil {
		tools.Exit("failed to write to ", allCountriesURI, err.Error())
	}
}

const (
	geonamesURL     = "http://download.geonames.org/export/dump/"
	commentSymbol   = byte('#')
	newLineSymbol   = byte('\n')
	delimSymbol     = byte('\t')
	boolTrue        = "1"
	allCountriesURI = `allCountries.zip`
)

func sParse(s *bufio.Scanner, headerLength uint, f func([]string) bool) {
	var err error
	var line string
	var rawSplit []string
	for s.Scan() {
		if headerLength != 0 {
			headerLength--
			continue
		}
		line = s.Text()
		if len(line) == 0 {
			continue
		}
		if line[0] == commentSymbol {
			continue
		}
		rawSplit = strings.Split(line, "\t")
		if !f(rawSplit) {
			break
		}
	}
	if err = s.Err(); err != nil {
		log.Fatal(err)
	}
}

func parse(data []byte, headerLength int, f func([][]byte) bool) {
	rawSplit := bytes.Split(data, []byte{newLineSymbol})
	var rawLineSplit [][]byte
	for i := range rawSplit {
		if headerLength != 0 {
			headerLength--
			continue
		}
		if len(rawSplit[i]) == 0 {
			continue
		}
		if rawSplit[i][0] == commentSymbol {
			continue
		}
		rawLineSplit = bytes.Split(rawSplit[i], []byte{delimSymbol})
		if !f(rawLineSplit) {
			break
		}
	}
}
