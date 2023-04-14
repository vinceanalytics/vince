package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"github.com/klauspost/compress/zstd"
)

var client = http.Client{}

func main() {
	download()
	err := processCountry()
	if err != nil {
		log.Fatal(err)
	}
}

func download() {
	if os.Getenv("DOWNLOAD") == "" {
		return
	}
	err := downloadCountries()
	if err != nil {
		log.Fatal(err)
	}
}

type Feature struct {
	ID   string
	Name string
}

func processCountry() error {
	r, err := zip.OpenReader(allCountriesURI)
	if err != nil {
		return err
	}
	defer r.Close()

	o, err := r.File[0].Open()
	if err != nil {
		return err
	}
	defer o.Close()

	db, err := badger.Open(badger.DefaultOptions("").WithInMemory(true))
	if err != nil {
		return err
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
				log.Fatal(err)
			}
		}
		return true
	})

	f, err := os.Create("city_geoname_db.zstd")
	if err != nil {
		return err
	}
	defer f.Close()
	e, _ := zstd.NewWriter(f, zstd.WithEncoderLevel(
		zstd.SpeedBestCompression,
	))
	_, err = db.Backup(e, 0)
	if err != nil {
		return err
	}
	e.Close()
	return nil
}

func downloadCountries() error {
	url := fmt.Sprintf("%s%s", geonamesURL, allCountriesURI)
	res, err := client.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	f, err := os.Create(allCountriesURI)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.ReadFrom(res.Body)
	return err
}

const (
	geonamesURL     = "http://download.geonames.org/export/dump/"
	commentSymbol   = byte('#')
	newLineSymbol   = byte('\n')
	delimSymbol     = byte('\t')
	boolTrue        = "1"
	allCountriesURI = `allCountries.zip`
)

func unzip(data []byte) ([]*zip.File, error) {
	var err error

	r, err := zip.NewReader(bytes.NewReader(data), (int64)(len(data)))
	if err != nil {
		return nil, err
	}

	return r.File, nil
}

func getZipData(files []*zip.File, name string) ([]byte, error) {
	var result []byte

	for _, f := range files {
		if f.Name == name {
			src, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer src.Close()

			result, err = ioutil.ReadAll(src)
			if err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}

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
