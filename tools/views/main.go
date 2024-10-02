package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"log"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/vince/internal/geo"
	"github.com/vinceanalytics/vince/internal/ref"
)

const (
	camera  = "Mozilla/5.0 (Linux; Android 5.0.2; DMC-CM1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3865.92 Mobile Safari/537.36"
	app     = "Mozilla/5.0 (Linux; Android 11; Pixel 3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Mobile Safari/537.36"
	desktop = "Mozilla/5.0 (X11; Arch Linux i686; rv:2.0) Gecko/20110321 Firefox/4.0"
	mobile  = "Pulse/4.0.5 (iPhone; iOS 7.0.6; Scale/2.00)"
)

//go:embed domains.json
var referrerData []byte
var refs []string

func init() {
	json.Unmarshal(referrerData, &refs)
}

func main() {
	err := view().Run(
		context.Background(),
		os.Args,
	)
	if err != nil {
		log.Fatal(err)
	}
}

func view() *cli.Command {
	return &cli.Command{
		Name:  "view",
		Usage: "sends pageview (only used for testing)",
		Flags: []cli.Flag{
			&cli.DurationFlag{
				Name:  "duration,d",
				Value: time.Hour,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			target := "http://localhost:8080/api/event"
			client := &http.Client{}
			tick := time.NewTicker(10 * time.Millisecond)
			done := time.NewTimer(c.Duration("duration"))
			setupReferral()
			for {
				select {
				case <-done.C:
					return nil
				case ts := <-tick.C:
					rq, err := request(target, ts.UTC().UnixMilli())
					if err != nil {
						return err
					}
					rs, err := client.Do(rq)
					if err != nil {
						return err
					}
					rs.Body.Close()
				}
			}

		},
	}
}

func date(ts time.Time) time.Time {
	yy, mm, dd := ts.Date()
	return time.Date(yy, mm, dd, 0, 0, 0, 0, time.UTC)
}

var (
	paths = []string{
		"/",
		"/login",
		"/settings",
		"/register",
		"/docs",
		"/docs/1",
		"/docs/2",
	}
	ips         = geo.Rand(10)
	agents      = []string{camera, app, desktop, mobile}
	eventsName  = []string{"pageview", "Outbound Link: Click", "Purchase"}
	utmCampaign = []string{"", "Referral", "Advertisement", "Email"}
	utmSource   = []string{"", "Facebook", "Twitter", "DuckDuckGo", "Google"}
	hostnames   = []string{"en.vinceanalytics.com", "es.vinceanalytics.com", "vinceanalytics.com"}
	referrals   map[string][]string
)

func setupReferral() {

	// add  random sources
	m := map[string]struct{}{}
	core := map[string]struct{}{}
	for _, v := range utmSource[:1] {
		m[v] = struct{}{}
		core[v] = struct{}{}
	}
	for range 4 {
		m[ref.Rand()] = struct{}{}
	}
	for k := range m {
		_, ok := core[k]
		if !ok {
			utmSource = append(utmSource, k)
		}
	}
	src := make([]string, 0, len(utmSource)-1)
	for _, v := range utmSource[1:] {
		src = append(src, strings.ToLower(v))
	}
	referrals = make(map[string][]string)

	for i := range refs {
		for j := range src {
			if strings.Contains(refs[i], src[j]) {
				referrals[src[j]] = append(referrals[src[j]], refs[i])
			}
		}
	}
}

func request(target string, ts int64) (*http.Request, error) {
	q := make(url.Values)
	src := randomString(utmSource)
	q.Set("utm_source", src)
	q.Set("utm_capmaign", randomString(utmCampaign))
	event := map[string]any{
		"name":     randomString(eventsName),
		"ts":       ts,
		"url":      "https://" + randomString(hostnames) + randomString(paths) + "?" + q.Encode(),
		"domain":   "vinceanalytics.com",
		"referrer": randomString(referrals[strings.ToLower(src)]),
	}
	data, _ := json.Marshal(event)
	req, err := http.NewRequest(http.MethodPost, target, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Client-IP", randomString(ips))
	req.Header.Set("user-agent", randomString(agents))
	return req, nil
}

func randomString(ls []string) string {
	if len(ls) == 0 {
		return ""
	}
	i := rand.IntN(len(ls))
	return ls[i]
}

func generateDates(days int, ts time.Time) []time.Time {
	o := make([]time.Time, days)
	for i := range o {
		o[i] = ts.AddDate(0, 0, -i)
	}
	slices.Reverse(o)
	return o
}
