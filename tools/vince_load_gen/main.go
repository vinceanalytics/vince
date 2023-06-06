package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"

	"github.com/dop251/goja"
	"github.com/urfave/cli/v3"
)

func main() {
	a := &cli.App{
		Name:  "load_gen",
		Usage: "generates web analytics events",
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
					UserAgent: GetUserAgent(),
					IP:        GetIP(),
					Domain:    "vince.io",
					Host:      "http://localhost:8080",
					Path:      "/",
					Event:     "pageviews",
					Referrer:  GetReferrer(),
				}
				a := vm.ToValue(s).(*goja.Object)
				a.SetPrototype(call.This.Prototype())
				return a
			}
			vm.Set("Session", create)
			vm.Set("ip", GetIP)
			vm.Set("referer", GetReferrer)
			vm.Set("userAgent", GetUserAgent)
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

type Session struct {
	UserAgent *UserAgent `json:"user_agent"`
	IP        string     `json:"ip"`
	Host      string     `json:"host"`
	Website   string     `json:"website"`
	Domain    string     `json:"domain"`
	Path      string     `json:"path"`
	Event     string     `json:"event"`
	Referrer  string     `json:"referer"`
}

func (s *Session) With(key, value string) *Session {
	o := *s
	a := *s.UserAgent
	o.UserAgent = &a
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

func (o *Session) RequestBody() *Request {
	return &Request{
		EventName:   o.Event,
		Domain:      o.Domain,
		Referrer:    o.Referrer,
		URI:         o.Website + o.Path,
		ScreenWidth: o.UserAgent.ScreenWidth,
	}
}

func (o *Session) Send() {
	b, _ := json.Marshal(o.RequestBody())
	r, _ := http.NewRequest(http.MethodPost, o.Host+"/api/event", bytes.NewReader(b))
	r.Header.Set("x-forwarded-for", o.IP)
	r.Header.Set("user-agent", o.UserAgent.UserAgent)
	r.Header.Set("content-type", "text/plain")
	res, err := client.Do(r)
	if err != nil {
		println("> failed sending request", err.Error())
		return
	}
	defer res.Body.Close()
	println(res.Status)
}
