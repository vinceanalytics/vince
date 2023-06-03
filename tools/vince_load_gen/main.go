package main

import (
	"bytes"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	o := &Options{}
	a := &cli.App{
		Name:  "load_gen",
		Usage: "generates web analytics events",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "domain,d",
				Value:       "vince.io",
				Destination: &o.Domain,
			},
			&cli.StringFlag{
				Name:        "host,h,d",
				Value:       "http://localhost:8080",
				Destination: &o.Host,
			},
			&cli.StringFlag{
				Name:        "path,h,p",
				Value:       "/",
				Destination: &o.Path,
			},
			&cli.StringFlag{
				Name:        "event,h,e",
				Value:       "pageviews",
				Destination: &o.Event,
			},
			&cli.StringFlag{
				Name:        "referrer,h,r",
				Value:       GetReferrer(),
				Destination: &o.Referrer,
			},
			&cli.StringFlag{
				Name:        "website,h,w",
				Destination: &o.Website,
			},
		},
		Action: func(ctx *cli.Context) error {
			a := GetUserAgent()
			ip := GetIP()
			b, _ := json.Marshal(Request{
				EventName:   o.Event,
				Domain:      o.Domain,
				Referrer:    o.Referrer,
				URI:         o.Website + o.Path,
				ScreenWidth: a.ScreenWidth,
			})
			println(string(b))
			r, _ := http.NewRequest(http.MethodPost, o.Host+"/api/event", bytes.NewReader(b))
			r.Header.Set("x-forwarded-for", ip)
			r.Header.Set("user-agent", a.UserAgent)
			r.Header.Set("content-type", "text/plain")
			res, err := client.Do(r)
			if err != nil {
				return err
			}
			s, _ := httputil.DumpResponse(res, true)
			println(string(s))
			return res.Body.Close()
		},
	}
	err := a.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

type Options struct {
	Host     string
	Website  string
	Domain   string
	Path     string
	Event    string
	Referrer string
}

var client = &http.Client{}

type Request struct {
	EventName   string `json:"n"`
	URI         string `json:"url"`
	Referrer    string `json:"r"`
	Domain      string `json:"d"`
	ScreenWidth int    `json:"w"`
}

func GetReferrer() string {
	return domains[rand.Intn(len(domains))]
}
