package referrer

import (
	"sort"
	"strings"
)

//go:generate go run gen/main.go

func ParseReferrer(host string) *Medium {
	host = strings.TrimPrefix(host, "www.")
	parts := strings.Split(host, ".")
	sort.Sort(sort.Reverse(stringSlice(parts)))
	if len(parts) > maxReferrerSize {
		parts = parts[:maxReferrerSize]
	}
	for i := len(parts); i >= minReferrerSize; i -= 1 {
		host = strings.Join(parts[:i], ".")
		if m, ok := refList[host]; ok {
			return &Medium{Type: m.Type, Name: m.Name}
		}
	}
	return nil
}

type stringSlice []string

func (x stringSlice) Len() int           { return len(x) }
func (x stringSlice) Less(i, j int) bool { return i < j }
func (x stringSlice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
