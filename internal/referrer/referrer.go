package referrer

import (
	"embed"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/vinceanalytics/vince/pkg/log"
)

//go:generate go run gen/main.go

//go:embed icons
var Icons embed.FS

func ParseReferrer(host string) string {
	host = strings.TrimPrefix(host, "www.")
	parts := strings.Split(host, ".")
	sort.Sort(sort.Reverse(stringSlice(parts)))
	if len(parts) > maxReferrerSize {
		parts = parts[:maxReferrerSize]
	}
	for i := len(parts); i >= minReferrerSize; i -= 1 {
		host = strings.Join(parts[:i], ".")
		if m, ok := refList[host]; ok {
			return m
		}
	}
	return ""
}

type stringSlice []string

func (x stringSlice) Len() int           { return len(x) }
func (x stringSlice) Less(i, j int) bool { return i < j }
func (x stringSlice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

func ServeFavicon(key string, w http.ResponseWriter, r *http.Request) bool {
	key = strings.ToLower(key)
	key = strings.Replace(key, " ", "-", -1)
	if _, ok := favicon[key]; !ok {
		return ok
	}
	file := "icons/" + key + ".ico"
	f, err := Icons.Open(file)
	if err != nil {
		log.Get().Err(err).Msg("failed getting icon file")
		return false
	}
	http.ServeContent(w, r, file, time.Unix(int64(modTime), 0).UTC(), f.(io.ReadSeeker))
	return true
}
