package referrer

import (
	"sort"
	"strings"
)

//go:generate go run gen/main.go

func Parse(host string) string {
	host = strings.TrimPrefix(host, "www.")
	parts := strings.Split(host, ".")
	sort.Sort(sort.Reverse(StringSlice(parts)))
	if len(parts) > maxReferrerSize {
		parts = parts[:maxReferrerSize]
	}
	for i := len(parts); i >= minReferrerSize; i -= 1 {
		host = strings.Join(parts[:i], ".")
		if m, ok := RefList[host]; ok {
			return m
		}
	}
	return ""
}

type StringSlice []string

func (x StringSlice) Len() int           { return len(x) }
func (x StringSlice) Less(i, j int) bool { return i < j }
func (x StringSlice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
