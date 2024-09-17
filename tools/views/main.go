package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"slices"
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
		Flags: []cli.Flag{},
		Action: func(ctx context.Context, c *cli.Command) error {
			target := "http://localhost:8080/api/event"
			today := date(time.Now().UTC())
			days := generateDates(5, today)
			client := &http.Client{}
			for i := range days {
				day := days[i]
				for n := range 500 {
					rq, err := request(target, day.Add(time.Duration(n)*time.Minute).UnixMilli())
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
			return nil
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
)

func request(target string, ts int64) (*http.Request, error) {
	q := make(url.Values)
	q.Set("utm_source", randomString(utmSource))
	q.Set("utm_capmaign", randomString(utmCampaign))
	event := map[string]any{
		"name":     randomString(eventsName),
		"ts":       ts,
		"url":      "https://" + randomString(hostnames) + randomString(paths) + "?" + q.Encode(),
		"domain":   "vinceanalytics.com",
		"referrer": ref.Rand(),
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
