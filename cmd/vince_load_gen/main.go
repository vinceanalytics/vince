package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

func main() {
	duration := flag.Duration("d", 10*time.Minute, "duration of running the script")
	requests := flag.Int("r", 1, "number of requests")
	host := flag.String("h", "http://localhost:8080", "vince analytics  instance")
	flag.Parse()
	domain := flag.Arg(0)
	ctx, _ := context.WithTimeout(context.Background(), *duration)
	s := BasicSession(domain)
	wg := &sync.WaitGroup{}
	for i := 0; i < *requests; i += 1 {
		wg.Add(1)
		err := s.Execute(ctx, *host)
		if err != nil {
			log.Println(err)
			return
		}
	}
	wg.Wait()
}

var client = &http.Client{}

type User struct {
	Agent  UserAgent
	URI    string
	IP     string
	Domain string
}

func (u *User) Journey(ctx context.Context, host string, journeys ...*Journey) error {
	t := time.NewTimer(time.Second)
	defer t.Stop()
	for _, j := range journeys {
		for _, e := range j.Events {
			r := u.Request(ctx, j.From, host, j.Path, e.Name)
			res, err := client.Do(r)
			if err != nil {
				return err
			}
			if res.StatusCode != http.StatusOK {
				res.Body.Close()
				return fmt.Errorf("expected 200 got %d", res.StatusCode)
			}
			res.Body.Close()
			t.Reset(e.Stay)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-t.C:
			}
		}
	}
	return nil
}

func (u *User) Request(ctx context.Context, from, host, path, event string) *http.Request {
	if from == "" {
		from = u.Domain
	}
	b, _ := json.Marshal(Request{
		EventName:   event,
		Domain:      u.Domain,
		Referrer:    from,
		URI:         u.URI + path,
		ScreenWidth: u.Agent.ScreenWidth,
	})
	r, _ := http.NewRequestWithContext(ctx, http.MethodPost, host+"/api/event", bytes.NewReader(b))
	r.Header.Set("x-forwarded-for", u.IP)
	r.Header.Set("user-agent", u.Agent.UserAgent)
	r.Header.Set("content-type", "text/plain")
	return r
}

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

type Journey struct {
	From   string
	Path   string
	Events []*Event
}

type Event struct {
	Name string
	Stay time.Duration
}

type Session struct {
	URI      string
	Domain   string
	Journeys []*Journey
}

func (s *Session) Execute(ctx context.Context, host string) error {
	u := &User{
		Agent:  GetUserAgent(),
		IP:     GetIP(),
		URI:    s.URI,
		Domain: s.Domain,
	}
	log.Println("executing session ", u.IP)
	return u.Journey(ctx, host, s.Journeys...)
}

// a simple user journey. A user comes from a random site and starts at / , stays
// for 1 ms and clicks /about stats 1 ms and leaves the site.
func BasicSession(domain string) Session {
	return Session{
		URI:    "https://vince.test",
		Domain: domain,
		Journeys: []*Journey{
			{Path: "/", From: GetReferrer(), Events: []*Event{
				{Name: "pageview", Stay: time.Millisecond},
			}},
			{Path: "/about", Events: []*Event{
				{Name: "pageview", Stay: time.Millisecond},
			}},
		},
	}
}
