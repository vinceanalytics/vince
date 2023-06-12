package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/dop251/goja"
	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/vince/pkg/entry"
)

func main() {
	a := &cli.App{
		Name:  "vlg",
		Usage: "generates web analytics events",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "host",
				Value:   "http://localhost:8080",
				EnvVars: []string{"VLG_HOST"},
			},
		},
		Action: func(ctx *cli.Context) error {
			vm := goja.New()
			b, err := os.ReadFile(ctx.Args().First())
			if err != nil {
				return err
			}
			vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
			vm.Set("println", fmt.Println)
			create := func(call goja.ConstructorCall) *goja.Object {
				s := &Session{
					UserAgent: ua(),
					IP:        ip(),
					Domain:    "vince.io",
					Host:      ctx.String("host"),
					Path:      "/",
					Event:     "pageview",
					Referrer:  referrer(),
				}
				a := vm.ToValue(s).(*goja.Object)
				a.SetPrototype(call.This.Prototype())
				return a
			}
			vm.Set("Session", create)
			vm.Set("ip", ip)
			vm.Set("referer", referrer)
			vm.Set("userAgent", ua)
			_, err = vm.RunString(string(b))
			if err != nil {
				return err
			}
			return nil
		},
	}
	err := a.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

var client = &http.Client{}

func referrer() string {
	return "https://" + domains[rand.Intn(len(domains))]
}

type Session struct {
	// When this is true send will not send http request buf instead will buffer
	// it, this is useful to generate fixtures to use in testing.
	Fixture   bool     `json:"fixture"`
	Requests  Requests `json:"requests"`
	UserAgent UA       `json:"user_agent"`
	IP        string   `json:"ip"`
	Host      string   `json:"host"`
	Website   string   `json:"website"`
	Domain    string   `json:"domain"`
	Path      string   `json:"path"`
	Event     string   `json:"event"`
	Referrer  string   `json:"referer"`
}

type Requests []*entry.Request

func (r Requests) Dump() string {
	b, _ := json.Marshal(r)
	return string(b)
}

func (s *Session) With(key, value string) *Session {
	o := *s
	if o.Fixture && len(o.Requests) > 0 {
		o.Requests = make(Requests, len(s.Requests))
		copy(o.Requests, s.Requests)
	}
	switch key {
	case "host":
		o.Host = value
	case "website":
		o.Website = value
	case "domain":
		o.Domain = value
	case "path":
		o.Path = value
	case "event":
		o.Event = value
	}
	return &o
}

func (o *Session) RequestBody() *entry.Request {
	r := entry.NewRequest()
	r.EventName = o.Event
	r.Domain = o.Domain
	r.Referrer = o.Referrer
	r.URI = o.Website + o.Path
	r.ScreenWidth = o.UserAgent.ScreenWidth
	if o.Fixture {
		r.IP = o.IP
		r.UserAgent = o.UserAgent.UserAgent
	}
	return r
}

func (o *Session) Send() *Session {
	rb := o.RequestBody()
	if o.Fixture {
		o.Requests = append(o.Requests, rb)
		return o
	}
	b, _ := json.Marshal(rb)
	r, _ := http.NewRequest(http.MethodPost, o.Host+"/api/event", bytes.NewReader(b))
	r.Header.Set("x-forwarded-for", o.IP)
	r.Header.Set("user-agent", o.UserAgent.UserAgent)
	r.Header.Set("content-type", "text/plain")
	res, err := client.Do(r)
	if err != nil {
		println("> failed sending request", err.Error())
		return o
	}
	defer res.Body.Close()
	println(res.Status)
	return o
}

func (s *Session) Wait(n int) *Session {
	time.Sleep(time.Millisecond * time.Duration(n))
	return s
}
