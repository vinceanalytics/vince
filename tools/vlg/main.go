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
	"golang.org/x/time/rate"
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
		Commands: []*cli.Command{
			{Name: "bench", Action: bench},
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
				if len(call.Arguments) > 0 {
					s.Fixture = true
				}
				a := vm.ToValue(s).(*goja.Object)
				a.SetPrototype(call.This.Prototype())
				return a
			}
			vm.Set("Session", create)
			vm.Set("ip", ip)
			vm.Set("referer", referrer)
			vm.Set("userAgent", ua)
			vm.Set("console", console{})
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

// sends 1 billion page view events
func bench(ctx *cli.Context) error {
	bound := 1 << 30
	r := rate.NewLimiter(rate.Every(time.Second), 200)
	n := 0
	s := newSession(ctx.String("host"))
	chunk := 2 << 10
	for n < bound {
		for x := 0; r.Allow() && x < chunk; x++ {
			n++
			s.Reset().Send()
		}
	}
	return nil
}

var paths = []string{"/", "/about", "/career"}

func newSession(host string) *Session {
	return &Session{
		UserAgent: ua(),
		IP:        ip(),
		Domain:    "vince.io",
		Host:      host,
		Path:      "/",
		Event:     "pageview",
		Referrer:  referrer(),
	}
}

func (s *Session) Reset() *Session {
	s.UserAgent = ua()
	s.IP = ip()
	s.Path = paths[rand.Intn(len(paths))]
	s.Referrer = referrer()
	return s
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

func (s *Session) Pretty() string {
	b, _ := json.MarshalIndent(s.Requests, "", " ")
	return string(b)
}

func (o *Session) With(key, value string) *Session {
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
	return o
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
	rb.Release()
	r, _ := http.NewRequest(http.MethodPost, o.Host+"/api/event", bytes.NewReader(b))
	r.Header.Set("x-forwarded-for", o.IP)
	r.Header.Set("user-agent", o.UserAgent.UserAgent)
	r.Header.Set("Accept", "application/json")
	r.Header.Set("content-type", "application/json")

	res, err := client.Do(r)
	if err != nil {
		println("> failed sending request", err.Error())
		return o
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		println(res.Status)
	}
	return o
}

func (s *Session) Wait(n int) *Session {
	time.Sleep(time.Millisecond * time.Duration(n))
	return s
}

type console struct{}

func (console) Log(a ...any) {
	fmt.Fprintln(os.Stdout, a...)
}
