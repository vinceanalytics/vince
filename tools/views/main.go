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
			start := today.AddDate(0, 0, -720)
			client := &http.Client{}
			for i := range 720 {
				day := start.AddDate(0, 0, i)
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
	utmCampaign = []string{"", "Referral", "Advertisement", "Email"}
	utmSource   = []string{"", "Facebook", "Twitter", "DuckDuckGo", "Google"}
	hostnames   = []string{"en.vinceanalytics.com", "es.vinceanalytics.com", "vinceanalytics.com"}
)

func request(target string, ts int64) (*http.Request, error) {
	q := make(url.Values)
	q.Set("utm_source", randomString(utmSource))
	q.Set("utm_capmaign", randomString(utmCampaign))
	event := map[string]any{
		"name":     "pageview",
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
