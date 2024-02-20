package curl

import (
	"bytes"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

var API = CMD("http://localhost:8080")

const siteId = "?site_id=vinceanalytics.com"

func TestVersion(t *testing.T) {
	check(t, false, "version.sh", "/api/v1/version", http.MethodGet, nil, nil)
}
func TestVisitors(t *testing.T) {
	check(t, true, "visitors.sh", "/api/v1/stats/realtime/visitors"+siteId, http.MethodGet, nil, nil)
}

func TestAggregate(t *testing.T) {
	q := make(url.Values)
	q.Set("site_id", "vinceanalytics.com")
	q.Set("metrics", "visitors,visits,pageviews,views_per_visit,bounce_rate,visit_duration,events")
	check(t, true, "aggregate.sh", "/api/v1/stats/aggregate?"+q.Encode(), http.MethodGet, nil, nil)
}
func TestBreakdown(t *testing.T) {
	q := make(url.Values)
	q.Set("site_id", "vinceanalytics.com")
	q.Set("property", "browser")
	q.Set("metrics", "visitors,bounce_rate")
	check(t, true, "breakdown.sh", "/api/v1/stats/breakdown?"+q.Encode(), http.MethodGet, nil, nil)
}

func TestTimeseries(t *testing.T) {
	q := make(url.Values)
	q.Set("site_id", "vinceanalytics.com")
	q.Set("period", "day")
	q.Set("interval", "minute")
	check(t, true, "timeseries.sh", "/api/v1/stats/timeseries?"+q.Encode(), http.MethodGet, nil, nil)
}

func check(t *testing.T, write bool, file string, path, method string, headers http.Header, body proto.Message) {
	t.Helper()
	file = filepath.Join("testdata/", file)
	var b bytes.Buffer
	err := API.Format(&b, path, method, headers, body)
	require.NoError(t, err)
	if write {
		os.WriteFile(file, b.Bytes(), 0600)
	}
	want, err := os.ReadFile(file)
	require.NoError(t, err)
	require.Equal(t, string(want), b.String())
}
