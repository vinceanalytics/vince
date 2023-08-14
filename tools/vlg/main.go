package main

import (
	"bytes"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/vince/internal/entry"
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
	UserAgent UA     `json:"user_agent"`
	IP        string `json:"ip"`
	Host      string `json:"host"`
	Website   string `json:"website"`
	Domain    string `json:"domain"`
	Path      string `json:"path"`
	Event     string `json:"event"`
	Referrer  string `json:"referer"`
}

func (o *Session) RequestBody() *entry.Request {
	r := entry.NewRequest()
	r.EventName = o.Event
	r.Domain = o.Domain
	r.Referrer = o.Referrer
	r.URI = o.Website + o.Path
	r.ScreenWidth = o.UserAgent.ScreenWidth
	return r
}

func (o *Session) Send() *Session {
	rb := o.RequestBody()
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
