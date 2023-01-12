package main

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"math/rand"
	"sort"
)

//go:embed user_agents.json.gz
var uaZip []byte

type UserAgent struct {
	AppName    string `json:"appName"`
	Connection struct {
		Downlink      float64 `json:"downlink"`
		EffectiveType string  `json:"effectiveType"`
		RTT           int     `json:"rtt"`
	} `json:"connection"`
	Platform       string  `json:"platform"`
	PluginsLength  int     `json:"pluginsLength"`
	Vendor         string  `json:"vendor"`
	UserAgent      string  `json:"userAgent"`
	ViewportHeight int     `json:"viewportHeight"`
	ViewportWidth  int     `json:"viewportWidth"`
	DeviceCategory string  `json:"deviceCategory"`
	ScreenHeight   int     `json:"screenHeight"`
	ScreenWidth    int     `json:"screenWidth"`
	Weight         float64 `json:"weight"`
}

var userAgents []UserAgent
var weightIndex []float64

func init() {
	r, err := gzip.NewReader(bytes.NewReader(uaZip))
	if err != nil {
		panic(err)
	}
	err = json.NewDecoder(r).Decode(&userAgents)
	if err != nil {
		panic(err)
	}
	var totalWeight float64
	for i := 0; i < len(userAgents); i++ {
		totalWeight += userAgents[i].Weight
	}
	var sum float64
	for i := 0; i < len(userAgents); i++ {
		sum += userAgents[i].Weight / totalWeight
		weightIndex = append(weightIndex, sum)
	}
}

func GetUserAgent() UserAgent {
	r := rand.Float64()
	idx := sort.Search(len(weightIndex), func(i int) bool {
		return weightIndex[i] > r
	})
	return userAgents[idx]
}
