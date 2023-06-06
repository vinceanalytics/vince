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

type UA struct {
	UserAgent   string
	Weight      float64
	ScreenWidth int
}

var agents []UA
var weightIndex []float64

func init() {
	r, err := gzip.NewReader(bytes.NewReader(uaZip))
	if err != nil {
		panic(err)
	}
	err = json.NewDecoder(r).Decode(&agents)
	if err != nil {
		panic(err)
	}
	var totalWeight float64
	for i := 0; i < len(agents); i++ {
		totalWeight += agents[i].Weight
	}
	var sum float64
	for i := 0; i < len(agents); i++ {
		sum += agents[i].Weight / totalWeight
		weightIndex = append(weightIndex, sum)
	}
}

func ua() UA {
	r := rand.Float64()
	idx := sort.Search(len(weightIndex), func(i int) bool {
		return weightIndex[i] > r
	})
	return agents[idx]
}
