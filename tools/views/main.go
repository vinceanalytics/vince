package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/urfave/cli/v3"
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
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "endpoint",
				Value: "http://localhost:8080",
			},
			&cli.StringFlag{
				Name:  "name",
				Value: "pageview",
			},
			&cli.StringFlag{
				Name:  "referrer",
				Value: "m.facebook.com",
			},
			&cli.StringFlag{
				Name:  "url",
				Value: "https://vinceanalytics.com",
			},
			&cli.StringFlag{
				Name:  "path",
				Value: "/",
			},
			&cli.StringFlag{
				Name:  "domain",
				Value: "vinceanalytics.com",
			},
			&cli.StringFlag{
				Name:  "ip",
				Value: "127.0.0.1",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			event := map[string]any{
				"name":     c.String("name"),
				"url":      c.String("url") + c.String("path"),
				"domain":   c.String("domain"),
				"referrer": c.String("referrer"),
			}
			data, _ := json.Marshal(event)
			req, err := http.NewRequest(http.MethodPost, c.String("endpoint")+"/api/event", bytes.NewReader(data))
			if err != nil {
				return err
			}
			req.Header.Set("X-Client-IP", c.String("ip"))
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer res.Body.Close()
			o, _ := httputil.DumpResponse(res, true)
			fmt.Println(string(o))
			return nil
		},
	}
}
